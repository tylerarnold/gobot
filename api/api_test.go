package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hybridgroup/gobot"
)

func initTestAPI() *api {
	log.SetOutput(gobot.NullReadWriteCloser{})
	g := gobot.NewGobot()
	a := NewAPI(g)
	a.start = func(m *api) {}
	a.Start()
	a.SetDebug()

	g.AddRobot(gobot.NewTestRobot("Robot1"))
	g.AddRobot(gobot.NewTestRobot("Robot2"))
	g.AddRobot(gobot.NewTestRobot("Robot3"))
	g.AddCommand("TestFunction", func(params map[string]interface{}) interface{} {
		message := params["message"].(string)
		return fmt.Sprintf("hey %v", message)
	})

	return a
}
func TestBasicAuth(t *testing.T) {
	a := initTestAPI()

	a.SetBasicAuth("admin", "password")

	request, _ := http.NewRequest("GET", "/", nil)
	request.SetBasicAuth("admin", "password")
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	request, _ = http.NewRequest("GET", "/", nil)
	request.SetBasicAuth("admin", "wrongPassword")
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 401)
}

func TestRobeaux(t *testing.T) {
	a := initTestAPI()
	// html assets
	request, _ := http.NewRequest("GET", "/index.html", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)
	// js assets
	request, _ = http.NewRequest("GET", "/js/app.js", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)
	// css assets
	request, _ = http.NewRequest("GET", "/css/style.css", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)
	// unknown asset
	request, _ = http.NewRequest("GET", "/js/fake/file.js", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 404)
}

func TestMcp(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, len(body), 2)
}

func TestMcpCommands(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/commands", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []string
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, []string{"TestFunction"})
}

func TestExecuteMcpCommand(t *testing.T) {
	var body interface{}
	a := initTestAPI()

	// known command
	request, _ := http.NewRequest("GET",
		"/commands/TestFunction",
		bytes.NewBufferString(`{"message":"Beep Boop"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "hey Beep Boop")

	// unknown command
	request, _ = http.NewRequest("GET",
		"/commands/TestFuntion1",
		bytes.NewBufferString(`{"message":"Beep Boop"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "Unknown Command")
}

func TestRobots(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/robots", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, len(body), 3)
}

func TestRobot(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/robots/Robot1", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body["name"].(string), "Robot1")
}

func TestRobotDevices(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/robots/Robot1/devices", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, len(body), 3)
}

func TestRobotCommands(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/robots/Robot1/commands", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []string
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, []string{"robotTestFunction"})
}

func TestExecuteRobotCommand(t *testing.T) {
	var body interface{}
	a := initTestAPI()
	// known command
	request, _ := http.NewRequest("GET",
		"/robots/Robot1/commands/robotTestFunction",
		bytes.NewBufferString(`{"message":"Beep Boop", "robot":"Robot1"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "hey Robot1, Beep Boop")

	// unknown command
	request, _ = http.NewRequest("GET",
		"/robots/Robot1/commands/robotTestFuntion1",
		bytes.NewBufferString(`{"message":"Beep Boop"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "Unknown Command")
}

func TestRobotDevice(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET",
		"/robots/Robot1/devices/Device1",
		nil,
	)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body["name"].(string), "Device1")
}

func TestRobotDeviceCommands(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET",
		"/robots/Robot1/devices/Device1/commands",
		nil,
	)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []string
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, len(body), 2)
}

func TestExecuteRobotDeviceCommand(t *testing.T) {
	var body interface{}
	a := initTestAPI()

	// known command
	request, _ := http.NewRequest("GET",
		"/robots/Robot1/devices/Device1/commands/TestDriverCommand",
		bytes.NewBufferString(`{"name":"human"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "hello human")

	// unknown command
	request, _ = http.NewRequest("GET",
		"/robots/Robot1/devices/Device1/commands/DriverCommand1",
		bytes.NewBufferString(`{"name":"human"}`),
	)
	request.Header.Add("Content-Type", "application/json")
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)

	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body, "Unknown Command")
}

func TestRobotConnections(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET", "/robots/Robot1/connections", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body []map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, len(body), 3)
}

func TestRobotConnection(t *testing.T) {
	a := initTestAPI()
	request, _ := http.NewRequest("GET",
		"/robots/Robot1/connections/Connection1",
		nil,
	)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)

	var body map[string]interface{}
	json.NewDecoder(response.Body).Decode(&body)
	gobot.Assert(t, body["name"].(string), "Connection1")
}

func TestAPIRouter(t *testing.T) {
	a := initTestAPI()

	a.Head("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ := http.NewRequest("HEAD", "/test", nil)
	response := httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	a.Get("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ = http.NewRequest("GET", "/test", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	a.Post("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ = http.NewRequest("POST", "/test", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	a.Put("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ = http.NewRequest("PUT", "/test", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	a.Delete("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ = http.NewRequest("DELETE", "/test", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)

	a.Options("/test", func(res http.ResponseWriter, req *http.Request) {})
	request, _ = http.NewRequest("OPTIONS", "/test", nil)
	response = httptest.NewRecorder()
	a.ServeHTTP(response, request)
	gobot.Assert(t, response.Code, 200)
}