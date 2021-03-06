package fs

import (
	"fmt"
	"github.com/mitchellh/packer/common/uuid"
	"io/ioutil"
	"os"
	"path/filepath"
)

const publicFileMode os.FileMode = 0644
const publicDirMode os.FileMode = 0755
const privateFileMode os.FileMode = 0600
const privateDirMode os.FileMode = 0700

const publicPath string = "public"
const privatePath string = "private"

type FsAPI struct {
	Id   string
	Path string
}

func NewAPI(path string, name string) (*FsAPI, error) {
	var fullPath string
	if len(name) > 0 {
		fullPath = filepath.Join(path, name)
	} else {
		fullPath = path
	}
	if err := os.MkdirAll(fullPath, publicDirMode); err != nil {
		return nil, fmt.Errorf("Could not create path '%s': %s", fullPath, err.Error())
	}
	return &FsAPI{Path: fullPath}, nil
}

func (fs *FsAPI) WriteLocal(name, content string) error {
	filename := filepath.Join(fs.Path, name)
	if err := ioutil.WriteFile(filename, []byte(content), privateFileMode); err != nil {
		return fmt.Errorf("Could not write file '%s': %s", filename, err.Error())
	}
	return nil
}

func (fs *FsAPI) ReadLocal(name string) (string, error) {
	filename := filepath.Join(fs.Path, name)
	if content, err := ioutil.ReadFile(filename); err != nil {
		return "", fmt.Errorf("Could not read file '%s': %s", filename, err.Error())
	} else {
		return string(content), nil
	}
}

func (fs *FsAPI) SendPublic(dstId string, name string, content string) error {
	path := filepath.Join(fs.Path, publicPath, dstId)
	if err := os.MkdirAll(path, publicDirMode); err != nil {
		return fmt.Errorf("Could not create path '%s': %s", path, err.Error())
	}
	filename := filepath.Join(path, name)
	if err := ioutil.WriteFile(filename, []byte(content), publicFileMode); err != nil {
		return fmt.Errorf("Could not write file '%s': %s", filename, err.Error())
	}
	return nil
}

func (fs *FsAPI) GetPublic(dstId string, name string) (string, error) {
	filename := filepath.Join(fs.Path, publicPath, dstId, name)
	if content, err := ioutil.ReadFile(filename); err != nil {
		return "", fmt.Errorf("Could not read file '%s': %s", filename, err.Error())
	} else {
		return string(content), nil
	}
}

func (fs *FsAPI) SendPrivate(dstId string, name string, content string) error {
	path := filepath.Join(fs.Path, privatePath, dstId)
	if err := os.MkdirAll(path, privateDirMode); err != nil {
		return fmt.Errorf("Could not create path '%s': %s", path, err.Error())
	}
	filename := filepath.Join(path, name)
	if err := ioutil.WriteFile(filename, []byte(content), privateFileMode); err != nil {
		return fmt.Errorf("Could not write file '%s': %s", filename, err.Error())
	}
	return nil
}

func (fs *FsAPI) GetPrivate(dstId string, name string) (string, error) {
	filename := filepath.Join(fs.Path, privatePath, dstId, name)
	if content, err := ioutil.ReadFile(filename); err != nil {
		return "", fmt.Errorf("Could not read file '%s': %s", filename, err.Error())
	} else {
		return string(content), nil
	}
}

func (fs *FsAPI) StorePublic(name string, content string) error {
	if len(fs.Id) == 0 {
		return fmt.Errorf("Id cannot be empty")
	}
	return fs.SendPublic(fs.Id, name, content)
}

func (fs *FsAPI) LoadPublic(name string) (string, error) {
	if len(fs.Id) == 0 {
		return "", fmt.Errorf("Id cannot be empty")
	}
	return fs.GetPublic(fs.Id, name)
}

func (fs *FsAPI) StorePrivate(name string, content string) error {
	if len(fs.Id) == 0 {
		return fmt.Errorf("Id cannot be empty")
	}
	return fs.SendPrivate(fs.Id, name, content)
}

func (fs *FsAPI) LoadPrivate(name string) (string, error) {
	if len(fs.Id) == 0 {
		return "", fmt.Errorf("Id cannot be empty")
	}
	return fs.GetPrivate(fs.Id, name)
}

func (fs *FsAPI) Push(dstId, name, queue, content string) error {
	path := filepath.Join(fs.Path, privatePath, dstId, name, queue)

	if err := os.MkdirAll(path, privateDirMode); err != nil {
		return fmt.Errorf("Could not create path '%s': %s", path, err.Error())
	}
	filename := filepath.Join(path, uuid.TimeOrderedUUID())
	if err := ioutil.WriteFile(filename, []byte(content), privateFileMode); err != nil {
		return fmt.Errorf("Could not write file '%s': %s", filename, err.Error())
	}
	return nil
}

func (fs *FsAPI) PushIncoming(dstId, queue, content string) error {
	return fs.Push(dstId, "incoming", queue, content)
}

func (fs *FsAPI) PushOutgoing(queue, content string) error {
	if 0 == len(fs.Id) {
		return fmt.Errorf("Id cannot be empty")
	}
	return fs.Push(fs.Id, "outgoing", queue, content)
}

func (fs *FsAPI) Pop(srcId, name, queue string) (string, error) {
	pattern := filepath.Join(fs.Path, privatePath, srcId, name, queue, "*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("Could not glob files: %s", err.Error())
	}

	if len(files) == 0 {
		return "", fmt.Errorf("Nothing to pop")
	}

	filename := files[0]
	if content, err := ioutil.ReadFile(filename); err != nil {
		return "", fmt.Errorf("Could not read file '%s': %s", filename, err.Error())
	} else {
		if err := os.Remove(filename); err != nil {
			return "", fmt.Errorf("Couldn't remove file: %s", err.Error())
		} else {
			return string(content), nil
		}
	}
}

func (fs *FsAPI) PopIncoming(queue string) (string, error) {
	if 0 == len(fs.Id) {
		return "", fmt.Errorf("Id cannot be empty")
	}
	return fs.Pop(fs.Id, "incoming", queue)
}

func (fs *FsAPI) PopOutgoing(srcId, queue string) (string, error) {
	return fs.Pop(srcId, "outgoing", queue)
}

func (fs *FsAPI) Size(id, name, queue string) (int, error) {
	pattern := filepath.Join(fs.Path, privatePath, id, name, queue, "*")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return 0, fmt.Errorf("Could not glob files: %s", err.Error())
	}
	return len(files), nil
}

func (fs *FsAPI) IncomingSize(queue string) (int, error) {
	if 0 == len(fs.Id) {
		return 0, fmt.Errorf("Id cannot be empty")
	}
	return fs.Size(fs.Id, "incoming", queue)
}

func (fs *FsAPI) OutgoingSize(id, queue string) (int, error) {
	return fs.Size(id, "outgoing", queue)
}
