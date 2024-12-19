package ollama

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

type Checker struct {
	OllamaURL         string
	ServerStartedByUs bool
}

func NewChecker(ollamaURL string) *Checker {
	return &Checker{
		OllamaURL:         ollamaURL,
		ServerStartedByUs: false,
	}
}

func (myChecker *Checker) CheckInstallation() bool {
	_, err := exec.LookPath("ollama")
	return err == nil
}

func (myChecker *Checker) InstallOllama() error {
	switch runtime.GOOS {
	case "darwin", "linux":
		fmt.Println("To install Ollama, run the following command in your terminal:")
		fmt.Println("curl https://ollama.ai/install.sh | sh")
	case "windows":
		fmt.Println("To install Ollama on Windows, please visit https://ollama.ai for installation instructions")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	return nil
}

func (myChecker *Checker) IsServerRunning() bool {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	_, err := client.Get(myChecker.OllamaURL)
	return err == nil
}

func (myChecker *Checker) StartServer() error {
	cmd := exec.Command("ollama", "serve")
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Ollama server: %v", err)
	}

	// Wait for the server to start
	for i := 0; i < 10; i++ {
		if myChecker.IsServerRunning() {
			myChecker.ServerStartedByUs = true
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("ollama server did not start within the expected time")
}

func (myChecker *Checker) StopServer() error {
	if !myChecker.ServerStartedByUs {
		return nil
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", "ollama.exe")
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to stop Ollama server: %v", err)
		}
	} else {
		cmd := exec.Command("pkill", "ollama")
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to stop Ollama server: %v", err)
		}
	}
	myChecker.ServerStartedByUs = false
	return nil
}
