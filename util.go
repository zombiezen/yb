package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

func ExecToStdout(cmdString string, targetDir string) error {
	fmt.Printf("Running: %s in %s\n", cmdString, targetDir)

	cmdArgs := strings.Fields(cmdString)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:len(cmdArgs)]...)
	cmd.Dir = targetDir
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	var stdoutBuf bytes.Buffer

	err := cmd.Start()

	if err != nil {
		return fmt.Errorf("cmd.Start() failed with '%s'\n", err)
	}

	go func() {
		//outputs := io.MultiWriter(os.Stdout, websock)
		outputs := io.MultiWriter(os.Stdout, &stdoutBuf)
		for {
			io.Copy(outputs, stdoutIn)
			io.Copy(outputs, stderrIn)
			time.Sleep(300 * time.Millisecond)
		}
	}()

	if err = cmd.Wait(); err != nil {
		fmt.Printf("Output:\n%s\n", stdoutBuf.String())
		fmt.Printf("Command failed with %s\n", err)

		return err
	}

	return nil

}

func PrependToPath(dir string) {
	currentPath := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s:%s", dir, currentPath)
	fmt.Printf("Setting PATH to %s\n", newPath)
	os.Setenv("PATH", newPath)
}

func ConfigFilePath(filename string) string {
	u, _ := user.Current()
	configDir := filepath.Join(u.HomeDir, ".config", "yb")
	MkdirAsNeeded(configDir)
	filePath := filepath.Join(configDir, filename)
	return filePath
}

func PathExists(path string) bool {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

func DirectoryExists(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return false
	}

	return true
}

func MkdirAsNeeded(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Making dir: %s\n", dir)
		if err := os.MkdirAll(dir, 0700); err != nil {
			fmt.Printf("Unable to create dir: %v\n", err)
			return err
		}
	}

	return nil
}

func TemplateToString(templateText string, data interface{}) (string, error) {
	t, err := template.New("generic").Parse(templateText)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		fmt.Printf("Can't render template:: %v\n", err)
		return "", err
	}

	result := tpl.String()
	return result, nil
}

func RemoveWritePermission(path string) bool {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false
	}

	//Check 'others' permission
	p := info.Mode()
	newmask := p & (0555)
	os.Chmod(path, newmask)

	return true
}

func RemoveWritePermissionRecursively(path string) bool {
	fileList := []string{}

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})

	if err != nil {
		return false
	}

	for _, file := range fileList {
		RemoveWritePermission(file)
	}

	return true
}
