package config

import (
    //"fmt"
    //"strings"
    "os"
    "io/ioutil"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConfigNew(t *testing.T) {
    conf := New("abc")
    assert.NotNil(t, conf)
    assert.Equal(t, conf.Path, "abc")
}

func TestConfigAddOrg(t *testing.T) {
    conf := New("abc")
    conf.AddOrg("hello", "somewhere")
    assert.Equal(t, conf.Data.Org[0].Name, "hello")
}

func TestConfigSaveLoad(t *testing.T) {
    file, _ := ioutil.TempFile(".", "xxx")
    filename := file.Name()
    file.Close()

    conf := New(filename)
    conf.AddOrg("org1", "path1")
    conf.AddOrg("org2", "path2")
    conf.Save()

    newConf := New(filename)
    err := newConf.Load()
    assert.Nil(t, err)
    assert.Equal(t, conf, newConf)
    os.Remove(filename)
}