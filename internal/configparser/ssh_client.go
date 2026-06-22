package configparser

import (
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient provides SSH connectivity for device configuration retrieval
type SSHClient struct {
	client *ssh.Client
	creds  SSHCredentials
}

// NewSSHClient creates a new SSH client with the given credentials
func NewSSHClient(creds SSHCredentials) *SSHClient {
	return &SSHClient{
		creds: creds,
	}
}

// Connect establishes SSH connection to the target device
func (c *SSHClient) Connect(host string) error {
	if c.creds.Port == 0 {
		c.creds.Port = 22
	}
	if c.creds.Timeout == 0 {
		c.creds.Timeout = 30 * time.Second
	}

	// Build SSH configuration
	config := &ssh.ClientConfig{
		User:            c.creds.Username,
		Timeout:         c.creds.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: In production, use proper host key verification
	}

	// Add authentication methods
	if c.creds.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(c.creds.Password))
	}

	if c.creds.PrivateKey != "" || c.creds.KeyFile != "" {
		signer, err := c.loadPrivateKey()
		if err != nil {
			return fmt.Errorf("failed to load private key: %w", err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	if len(config.Auth) == 0 {
		return fmt.Errorf("no authentication methods configured")
	}

	// Connect to the remote server
	address := fmt.Sprintf("%s:%d", host, c.creds.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	c.client = client
	return nil
}

// Execute runs a command on the remote device and returns the output
func (c *SSHClient) Execute(command string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

// ExecuteInteractive runs a command that may require interactive responses
func (c *SSHClient) ExecuteInteractive(
	command string,
	responses map[string]string,
) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := session.Start(command); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Handle interactive prompts
	go func() {
		defer stdin.Close()
		for _, response := range responses {
			// This is a simplified implementation
			// In a real implementation, you would read from stdout/stderr
			// and respond to specific prompts
			time.Sleep(1 * time.Second)
			fmt.Fprintf(stdin, "%s\n", response)
		}
	}()

	// Read all output
	var output strings.Builder
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			output.Write(buf[:n])
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				break
			}
			output.Write(buf[:n])
		}
	}()

	if err := session.Wait(); err != nil {
		return output.String(), fmt.Errorf("command execution failed: %w", err)
	}

	return output.String(), nil
}

// Close closes the SSH connection
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// IsConnected returns true if the SSH client is connected
func (c *SSHClient) IsConnected() bool {
	if c.client == nil {
		return false
	}

	// Test the connection with a simple command
	_, _, err := c.client.Conn.SendRequest("keepalive@openssh.com", true, nil)
	return err == nil
}

// TestConnection tests SSH connectivity without establishing a persistent connection
func TestConnection(host string, creds SSHCredentials) error {
	client := NewSSHClient(creds)
	if err := client.Connect(host); err != nil {
		return err
	}
	defer client.Close()

	// Try to execute a simple command to verify the connection works
	_, err := client.Execute("echo 'test'")
	return err
}

// GetDeviceInfo retrieves basic device information via SSH
func (c *SSHClient) GetDeviceInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Try common commands to get device information
	commands := map[string]string{
		"hostname": "hostname",
		"uptime":   "uptime",
		"uname":    "uname -a",
		"date":     "date",
	}

	for key, cmd := range commands {
		output, err := c.Execute(cmd)
		if err == nil {
			info[key] = strings.TrimSpace(output)
		}
	}

	return info, nil
}

// loadPrivateKey loads and parses the SSH private key, preferring in-memory PEM
// content (c.creds.PrivateKey) over a key file path (c.creds.KeyFile).
func (c *SSHClient) loadPrivateKey() (ssh.Signer, error) {
	var keyBytes []byte
	var err error
	if c.creds.PrivateKey != "" {
		keyBytes = []byte(c.creds.PrivateKey)
	} else {
		keyBytes, err = ioutil.ReadFile(c.creds.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
	}

	var signer ssh.Signer
	if c.creds.KeyPassphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(c.creds.KeyPassphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return signer, nil
}

// SSHClientPool manages multiple SSH connections for concurrent operations
type SSHClientPool struct {
	clients map[string]*SSHClient
	creds   map[string]SSHCredentials
}

// NewSSHClientPool creates a new SSH client pool
func NewSSHClientPool() *SSHClientPool {
	return &SSHClientPool{
		clients: make(map[string]*SSHClient),
		creds:   make(map[string]SSHCredentials),
	}
}

// AddHost adds a host to the pool with its credentials
func (p *SSHClientPool) AddHost(host string, creds SSHCredentials) {
	p.creds[host] = creds
}

// GetClient returns an SSH client for the specified host, creating one if needed
func (p *SSHClientPool) GetClient(host string) (*SSHClient, error) {
	client, exists := p.clients[host]
	if exists && client.IsConnected() {
		return client, nil
	}

	creds, hasCreds := p.creds[host]
	if !hasCreds {
		return nil, fmt.Errorf("no credentials configured for host %s", host)
	}

	client = NewSSHClient(creds)
	if err := client.Connect(host); err != nil {
		return nil, err
	}

	p.clients[host] = client
	return client, nil
}

// ExecuteOnHost executes a command on the specified host
func (p *SSHClientPool) ExecuteOnHost(host, command string) (string, error) {
	client, err := p.GetClient(host)
	if err != nil {
		return "", err
	}

	return client.Execute(command)
}

// CloseAll closes all SSH connections in the pool
func (p *SSHClientPool) CloseAll() {
	for _, client := range p.clients {
		client.Close()
	}
	p.clients = make(map[string]*SSHClient)
}

// ValidateHost checks if a host is reachable via SSH
func ValidateHost(host string, creds SSHCredentials) error {
	// First check if the host is reachable on the SSH port
	address := fmt.Sprintf("%s:%d", host, creds.Port)
	if creds.Port == 0 {
		address = fmt.Sprintf("%s:22", host)
	}

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("host %s is not reachable on SSH port: %w", host, err)
	}
	conn.Close()

	// Then try to establish an SSH connection
	return TestConnection(host, creds)
}

