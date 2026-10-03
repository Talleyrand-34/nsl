// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package configparser

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// defaultSSHPort is used whenever SSHCredentials.Port is left at zero.
const defaultSSHPort = 22

// sshAddress builds the dial address for host, defaulting the port to 22.
//
// net.JoinHostPort, not fmt.Sprintf("%s:%d", ...): an IPv6 host must be bracketed.
// Formatting produces "::1:22", where the colons of the literal cannot be told from
// the port separator, and every dialler rejects it ("too many colons in address").
// JoinHostPort yields "[::1]:22".
func sshAddress(host string, port int) string {
	if port == 0 {
		port = defaultSSHPort
	}
	return net.JoinHostPort(host, strconv.Itoa(port))
}

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
	// The default port is applied by sshAddress, so it is not restated here.
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
	address := sshAddress(host, c.creds.Port)
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

// Close closes the SSH connection
func (c *SSHClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
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

// ValidateHost checks if a host is reachable via SSH
func ValidateHost(host string, creds SSHCredentials) error {
	// First check if the host is reachable on the SSH port
	address := sshAddress(host, creds.Port)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("host %s is not reachable on SSH port: %w", host, err)
	}
	conn.Close()

	// Then try to establish an SSH connection
	return TestConnection(host, creds)
}
