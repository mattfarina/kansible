/*
 * Copyright 2016 Red Hat
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/fabric8io/kansible/log"
	"golang.org/x/crypto/ssh"
)

// RemoteSSHCommand invokes the given command on a host and port
func RemoteSSHCommand(user string, privateKey string, host string, port string, cmd string, envVars map[string]string) error {
	if len(privateKey) == 0 {
		return fmt.Errorf("Could not find PrivateKey for entry %s", host)
	}
	log.Info("Connecting to host over SSH on host %s and port %d with user %s with command `%s`", host, port, user, cmd)

	hostPort := net.JoinHostPort(host, port)

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			PublicKeyFile(privateKey),
		},
	}
	if sshConfig == nil {
		log.Warn("No sshConfig could be created!")
	}
	connection, err := ssh.Dial("tcp", hostPort, sshConfig)
	if err != nil {
		return fmt.Errorf("Failed to dial: %s", err)
	}
	session, err := connection.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		// ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("Request for pseudo terminal failed: %s", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdin for session: %v", err)
	}
	go io.Copy(stdin, os.Stdin)

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stdout for session: %v", err)
	}
	go io.Copy(os.Stdout, stdout)

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("Unable to setup stderr for session: %v", err)
	}
	go io.Copy(os.Stderr, stderr)

	for envName, envValue := range envVars {
		log.Info("Setting environment value %s = %s", envName, envValue)
		if err := session.Setenv(envName, envValue); err != nil {
			return fmt.Errorf("Could not set environment variable %s = %s over SSH. This could be disabled by the sshd configuration. See the `AcceptEnv` setting in your /etc/ssh/sshd_config more info: http://linux.die.net/man/5/sshd_config . Error: %s", envName, envValue, err)
		}
	}

	log.Info("Running command %s", cmd)
	err = session.Run(cmd)
	if err != nil {
		return fmt.Errorf("Failed to run command: "+cmd+": %v", err)
	}
	return nil
}

// PublicKeyFile creates the auth method for the given private key file
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}
