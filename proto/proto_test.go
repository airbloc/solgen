package proto

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Flaque/filet"
	"github.com/frostornge/solgen/deployment"
	"github.com/stretchr/testify/assert"
)

const TestDeployment = `{"Migrations":{"abi":[{"constant":true,"inputs":[],"name":"last_completed_migration","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[{"name":"completed","type":"uint256"}],"name":"setCompleted","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"new_address","type":"address"}],"name":"upgrade","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]}}`
const TestDeploymentError = `{"Migrations":{"abi":[{"constant":"Hello","inputs":[],"name":"last_completed_migration","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"owner","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"inputs":[],"payable":false,"stateMutability":"nonpayable","type":"constructor"},{"constant":false,"inputs":[{"name":"completed","type":"uint256"}],"name":"setCompleted","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"new_address","type":"address"}],"name":"upgrade","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]}}`
const TestDeploymentPath = "deployment.json"

func TestGenerateBind(t *testing.T) {
	defer filet.CleanUp(t)
	dirName := filet.TmpDir(t, "")
	tmpName := filepath.Join(dirName, TestDeploymentPath)
	filet.File(t, tmpName, TestDeployment)

	assert.NoError(t, os.Chdir(".."))

	deployments, err := deployment.GetDeploymentsFrom(tmpName)
	assert.NoError(t, err)
	assert.NoError(t, GenerateBind(dirName, deployments, Options{}))

	assert.True(t, filet.Exists(t, filepath.Join(dirName, "migrations.proto")))
}

func TestGenerateBind_Airbloc(t *testing.T) {
	assert.NoError(t, os.Chdir(".."))
	assert.NoError(t, os.RemoveAll("./test/proto"))

	deployments, err := deployment.GetDeploymentsFrom("http://localhost:8500")
	assert.NoError(t, err)
	assert.NoError(t, GenerateBind("./test/proto", deployments, Options{
		"Accounts":           {"(address,uint8,address,address)": "Account"},
		"AppRegistry":        {"(string,address,bytes32)": "App"},
		"ControllerRegistry": {"(address,uint256)": "DataController"},
		"DataTypeRegistry":   {"(string,address,bytes32)": "DataType"},
		"Exchange":           {"(string,address,bytes20[],uint256,uint256,(address,bytes4,bytes),uint8)": "Offer"},
	}))

}
