package devices

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jritsema/go-htmx-starter/internal"
	"github.com/jritsema/go-htmx-starter/pkg/templates"
	"github.com/jritsema/go-htmx-starter/pkg/webtools"
	"github.com/jritsema/gotoolbox/web"
	"go.etcd.io/bbolt"
)

// Delete -> DELETE /company/{id} -> delete, companys.html

// Edit   -> GET /company/edit/{id} -> row-edit.html
// Save   ->   PUT /company/{id} -> update, row.html
// Cancel ->	 GET /company/{id} -> nothing, row.html

// Add    -> GET /company/add/ -> companys-add.html (target body with row-add.html and row.html)
// Save   ->   POST /company -> add, companys.html (target body without row-add.html)
// Cancel ->	 GET /company -> nothing, companys.html

type DeviceThing struct {
	db *bbolt.DB
	//parsed templates
	html *template.Template
}

func NewDevices(db *bbolt.DB, router *http.ServeMux) DeviceThing {
	//parse templates
	var err error
	html, err := templates.TemplateParseFSRecursive(internal.TemplateFS, ".html", true, nil)
	if err != nil {
		panic(err)
	}

	dt := DeviceThing{
		db:   db,
		html: html,
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Devices"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}
	router.Handle("/device/add", web.Action(dt.DeviceAdd))
	router.Handle("/device/add/", web.Action(dt.DeviceAdd))

	router.Handle("/device/edit", web.Action(dt.DeviceEdit))
	router.Handle("/device/edit/", web.Action(dt.DeviceEdit))

	router.Handle("/device", web.Action(dt.Devices))
	router.Handle("/device/", web.Action(dt.Devices))

	router.Handle("/devices", web.Action(dt.Index))

	router.Handle("/device/connect", web.Action(dt.DeviceConnect))
	router.Handle("/device/connect/", web.Action(dt.DeviceConnect))

	return dt
}
func (dt DeviceThing) Index(r *http.Request) *web.Response {
	return webtools.HTML(r, http.StatusOK, dt.html, "devices/index.html", dt.GetDevices(), nil)
}

// GET /device/add
func (dt DeviceThing) DeviceAdd(r *http.Request) *web.Response {
	return webtools.HTML(r, http.StatusOK, dt.html, "devices/devices-add.html", dt.GetDevices(), nil)
}

// /GET company/edit/{id}
func (dt DeviceThing) DeviceEdit(r *http.Request) *web.Response {
	id, _ := web.PathLast(r)
	row := dt.GetDeviceByID(id)
	return webtools.HTML(r, http.StatusOK, dt.html, "devices/row-edit.html", row, nil)
}

// Connect to device
func (dt DeviceThing) DeviceConnect(r *http.Request) *web.Response {
	id, _ := web.PathLast(r)
	_ = dt.GetDeviceByID(id)
	// amt := amt.NewMessages(device.Address, device.Username, device.Password, true, false)
	// result, _ := amt.GeneralSettings.Get()
	// result.Body.AMTGeneralSettings
	return webtools.HTML(r, http.StatusOK, dt.html, "devices/device.html", nil, nil)
}

// GET /company
// GET /company/{id}
// DELETE /company/{id}
// PUT /company/{id}
// POST /company
func (dt DeviceThing) Devices(r *http.Request) *web.Response {
	id, segments := web.PathLast(r)
	switch r.Method {

	case http.MethodDelete:
		dt.DeleteDevice(id)
		return webtools.HTML(r, http.StatusOK, dt.html, "devices/devices.html", dt.GetDevices(), nil)

	//cancel
	case http.MethodGet:
		if segments > 1 {
			//cancel edit
			row := dt.GetDeviceByID(id)
			return webtools.HTML(r, http.StatusOK, dt.html, "devices/row.html", row, nil)
		} else {
			//cancel add
			return webtools.HTML(r, http.StatusOK, dt.html, "devices/devices.html", dt.GetDevices(), nil)
		}

	//save edit
	case http.MethodPut:
		row := dt.GetDeviceByID(id)
		r.ParseForm()
		row.Id, _ = strconv.Atoi(id)
		row.Name = r.Form.Get("name")
		row.Address = r.Form.Get("address")
		row.Username = r.Form.Get("username")
		row.Password = r.Form.Get("password")
		if !row.IsValid() {
			return webtools.HTML(r, http.StatusBadRequest, dt.html, "devices/errors.html", row, nil)
		}
		dt.UpdateDevice(row)
		return webtools.HTML(r, http.StatusOK, dt.html, "devices/row.html", row, nil)

	//save add
	case http.MethodPost:
		row := Device{}
		r.ParseForm()
		row.Id, _ = strconv.Atoi(r.Form.Get("id"))
		row.Name = r.Form.Get("name")
		row.Address = r.Form.Get("address")
		row.Username = r.Form.Get("username")
		row.Password = r.Form.Get("password")
		if !row.IsValid() {
			return webtools.HTML(r, http.StatusBadRequest, dt.html, "devices/errors.html", dt.GetDevices(), nil)
		}
		dt.AddDevice(row)
		return webtools.HTML(r, http.StatusOK, dt.html, "devices/devices.html", dt.GetDevices(), nil)
	}

	return web.Empty(http.StatusNotImplemented)
}
