package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	power "github.com/open-amt-cloud-toolkit/go-wsman-messages/v2/pkg/wsman/cim/power"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/open-amt-cloud-toolkit/console/internal/entity/dto"
	"github.com/open-amt-cloud-toolkit/console/pkg/logger"
)

var ErrGeneral = errors.New("general error")

func deviceManagementTest(t *testing.T) (*MockDeviceManagementFeature, *gin.Engine) {
	t.Helper()

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	log := logger.New("error")
	deviceManagement := NewMockDeviceManagementFeature(mockCtl)
	amtExplorerMock := NewMockAMTExplorerFeature(mockCtl)
	engine := gin.New()
	handler := engine.Group("/api/v1")

	NewAmtRoutes(handler, deviceManagement, amtExplorerMock, log)

	return deviceManagement, engine
}

var aGoodTime = time.Unix(int64(1073007983), 0)

func TestDeviceManagement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		url          string
		mock         func(m *MockDeviceManagementFeature)
		method       string
		requestBody  interface{}
		expectedCode int
		response     interface{}
	}{
		{
			name:   "getVersion - successful retrieval",
			url:    "/api/v1/amt/version/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetVersion(context.Background(), "valid-guid").
					Return(map[string]interface{}{"version": "1.0.0"}, nil)
			},
			expectedCode: http.StatusOK,
			response:     map[string]interface{}{"version": "1.0.0"},
		},
		{
			name:   "getFeatures - successful retrieval",
			url:    "/api/v1/amt/features/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetFeatures(context.Background(), "valid-guid").
					Return(dto.Features{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     map[string]interface{}{"IDER": false, "KVM": false, "SOL": false, "redirection": false, "optInState": 0, "userConsent": ""},
		},
		{
			name:        "setFeatures - successful setting",
			url:         "/api/v1/amt/features/valid-guid",
			method:      http.MethodPost,
			requestBody: dto.Features{},
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().SetFeatures(context.Background(), "valid-guid", dto.Features{}).Return(dto.Features{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     dto.Features{},
		},
		{
			name:   "getAlarmOccurrences - successful retrieval",
			url:    "/api/v1/amt/alarmOccurrences/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetAlarmOccurrences(context.Background(), "valid-guid").
					Return([]dto.AlarmClockOccurrence{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     []dto.AlarmClockOccurrence{},
		},
		{
			name:   "deleteAlarmOccurrences - successful deletion",
			url:    "/api/v1/amt/alarmOccurrences/valid-guid",
			method: http.MethodDelete,
			requestBody: dto.AlarmClockOccurrence{
				ElementName:        "elementName",
				StartTime:          aGoodTime,
				Interval:           1,
				DeleteOnCompletion: true,
				InstanceID:         "1",
			},
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().DeleteAlarmOccurrences(context.Background(), "valid-guid", "1").Return(nil)
			},
			expectedCode: http.StatusNoContent,
		},
		{
			name:   "getHardwareInfo - successful retrieval",
			url:    "/api/v1/amt/hardwareInfo/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetHardwareInfo(context.Background(), "valid-guid").
					Return(map[string]interface{}{"hardware": "info"}, nil)
			},
			expectedCode: http.StatusOK,
			response:     map[string]interface{}{"hardware": "info"},
		},
		{
			name:   "getDiskInfo - successful retrieval",
			url:    "/api/v1/amt/diskInfo/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetDiskInfo(context.Background(), "valid-guid").
					Return(map[string]interface{}{"disk": "info"}, nil)
			},
			expectedCode: http.StatusOK,
			response:     map[string]interface{}{"disk": "info"},
		},
		{
			name:   "getPowerState - successful retrieval",
			url:    "/api/v1/amt/power/state/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetPowerState(context.Background(), "valid-guid").
					Return(map[string]interface{}{"state": "on"}, nil)
			},
			expectedCode: http.StatusOK,
			response:     map[string]interface{}{"state": "on"},
		},
		{
			name:   "powerAction - successful action",
			url:    "/api/v1/amt/power/action/valid-guid",
			method: http.MethodPost,
			requestBody: dto.PowerAction{
				Action: 4,
			},
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().SendPowerAction(context.Background(), "valid-guid", 4).
					Return(power.PowerActionResponse{ReturnValue: 0}, nil)
			},
			expectedCode: http.StatusOK,
			response:     power.PowerActionResponse{ReturnValue: 0},
		},
		{
			name:   "getAuditLog - successful retrieval",
			url:    "/api/v1/amt/log/audit/valid-guid?startIndex=0",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetAuditLog(context.Background(), 0, "valid-guid").
					Return(dto.AuditLog{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     dto.AuditLog{},
		},
		{
			name:   "getEventLog - successful retrieval",
			url:    "/api/v1/amt/log/event/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetEventLog(context.Background(), "valid-guid").
					Return([]dto.EventLog{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     []dto.EventLog{},
		},
		{
			name:   "setBootOptions - successful setting",
			url:    "/api/v1/amt/power/bootOptions/valid-guid",
			method: http.MethodPost,
			requestBody: dto.BootSetting{
				Action: 109,
			},
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().SetBootOptions(context.Background(), "valid-guid", dto.BootSetting{
					Action: 109,
				}).Return(power.PowerActionResponse{ReturnValue: 0}, nil)
			},
			expectedCode: http.StatusOK,
			response:     power.PowerActionResponse{ReturnValue: 0},
		},
		{
			name:   "successful retrieval",
			url:    "/api/v1/amt/networkSettings/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetNetworkSettings(context.Background(), "valid-guid").
					Return(dto.NetworkSettings{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     dto.NetworkSettings{},
		},
		{
			name:   "getCertificates - successful retrieval",
			url:    "/api/v1/amt/certificates/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetCertificates(context.Background(), "valid-guid").
					Return(dto.SecuritySettings{}, nil)
			},
			expectedCode: http.StatusOK,
			response:     dto.SecuritySettings{},
		},
		{
			name:   "getCertificates - failed retrieval",
			url:    "/api/v1/amt/certificates/valid-guid",
			method: http.MethodGet,
			mock: func(m *MockDeviceManagementFeature) {
				m.EXPECT().GetCertificates(context.Background(), "valid-guid").
					Return(dto.SecuritySettings{}, ErrGeneral)
			},
			expectedCode: http.StatusInternalServerError,
			response:     dto.SecuritySettings{},
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deviceManagement, engine := deviceManagementTest(t)

			tc.mock(deviceManagement)

			var req *http.Request

			var err error

			if tc.method == http.MethodPost || tc.method == http.MethodPatch || tc.method == http.MethodDelete {
				reqBody, _ := json.Marshal(tc.requestBody)
				req, err = http.NewRequest(tc.method, tc.url, bytes.NewBuffer(reqBody))
			} else {
				req, err = http.NewRequest(tc.method, tc.url, http.NoBody)
			}

			if err != nil {
				t.Fatalf("Couldn't create request: %v\n", err)
			}

			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			require.Equal(t, tc.expectedCode, w.Code)

			if tc.expectedCode == http.StatusOK || tc.expectedCode == http.StatusCreated {
				jsonBytes, _ := json.Marshal(tc.response)
				require.Equal(t, string(jsonBytes), w.Body.String())
			}
		})
	}
}
