package deployment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/Flaque/filet"

	"github.com/stretchr/testify/assert"
)

var testContractNames = []string{
	"Accounts",
	"AppRegistry",
	"Consents",
	"ControllerRegistry",
	"DataTypeRegistry",
	"ERC20Escrow",
	"Exchange",
}

const TestDeployment = `{"Migrations":{"abi":[{"constant":true,"inputs":[],"name":"last_completed_migration","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[{"name":"completed","type":"uint256"}],"name":"setCompleted","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"new_address","type":"address"}],"name":"upgrade","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]}}`
const TestDeploymentError = `{"Migrations":{"abi":[{"constant":"Hello","inputs":[],"name":"last_completed_migration","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[{"name":"completed","type":"uint256"}],"name":"setCompleted","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"new_address","type":"address"}],"name":"upgrade","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]}}`
const TestDeploymentPath = "deployment.json"

func TestGetDeploymentsFromUrl(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write([]byte(TestDeployment)); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}),
	)

	deployments, err := GetDeploymentsFrom(testServer.URL)
	assert.NoError(t, err)

	_, ok := deployments["Migrations"]
	assert.True(t, ok)
}

func TestGetDeploymentsFromUrlInvalidPath(t *testing.T) {
	_, err := GetDeploymentsFrom("http://localhost")
	assert.IsType(t, &url.Error{}, err)
}

func TestGetDeploymentsFromUrlInvalidJson(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write([]byte("{" + TestDeploymentError)); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}),
	)

	_, err := GetDeploymentsFrom(testServer.URL)
	assert.IsType(t, &json.SyntaxError{}, err)
}

func TestGetDeploymentsFromUrlInvalidDeployment(t *testing.T) {
	testServer := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if _, err := writer.Write([]byte(TestDeploymentError)); err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
			}
		}),
	)

	_, err := GetDeploymentsFrom(testServer.URL)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)
}

func TestGetDeploymentsFromFile(t *testing.T) {
	defer filet.CleanUp(t)
	filet.File(t, TestDeploymentPath, TestDeployment)

	deployments, err := GetDeploymentsFrom(TestDeploymentPath)
	assert.NoError(t, err)

	_, ok := deployments["Migrations"]
	assert.True(t, ok)
}

func TestGetDeploymentsFromFileInvalidPath(t *testing.T) {
	_, err := GetDeploymentsFrom("deployment.local.json")
	assert.IsType(t, &os.PathError{}, err)
}

func TestGetDeploymentsFromFileInvalidJson(t *testing.T) {
	defer filet.CleanUp(t)
	filet.File(t, TestDeploymentPath, "{"+TestDeploymentError)

	_, err := GetDeploymentsFrom(TestDeploymentPath)
	assert.IsType(t, &json.SyntaxError{}, err)
}

func TestGetDeploymentsFromFileInvalidDeployment(t *testing.T) {
	defer filet.CleanUp(t)
	filet.File(t, TestDeploymentPath, TestDeploymentError)

	_, err := GetDeploymentsFrom(TestDeploymentPath)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)
}
