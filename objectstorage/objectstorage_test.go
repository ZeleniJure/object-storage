package objectstorage

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestParseId(t *testing.T) {
	o := &ObjectStorage{}
	const variable = "a1"
	req, err := http.NewRequest("GET", "/object/" + variable, nil)
	assert.NoError(t, err)

	vars := map[string]string{
        "id": variable,
    }

    // CHANGE THIS LINE!!!
    req = mux.SetURLVars(req, vars)

	id, err := o.parseId(req)
	assert.NoError(t, err)
	assert.Equal(t, id, "a1")
}

