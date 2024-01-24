// Copyright 2023 XigXog
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/xigxog/kubefox/api"
	"github.com/xigxog/kubefox/logkf"
	"github.com/xigxog/kubefox/utils"
	"github.com/xigxog/kubefox/vault"
)

const (
	TenYears = "87600h"
)

type Flags struct {
	instance          string
	platformNamespace string
	compName          string
	compSvcName       string
	compIP            string
	vaultURL          string
	logFormat         string
	logLevel          string
}

var (
	f Flags
)

func main() {
	flag.StringVar(&f.instance, "instance", "", "Name of KubeFox Instance. (required)")
	flag.StringVar(&f.platformNamespace, "platform-namespace", "", "Kubernetes Namespace of Platform. (required)")
	flag.StringVar(&f.compName, "component", "", "Name of Component to bootstrap. (required)")
	flag.StringVar(&f.compSvcName, "component-service-name", "", "Service name of Component to bootstrap. (required)")
	flag.StringVar(&f.compIP, "component-ip", "", "IP address of Component to bootstrap. (required)")
	flag.StringVar(&f.vaultURL, "vault-url", "", "URL of Vault server. (required)")
	flag.StringVar(&f.logFormat, "log-format", "console", `Log format; one of ["json", "console"].`)
	flag.StringVar(&f.logLevel, "log-level", "debug", `Log level; one of ["debug", "info", "warn", "error"].`)
	flag.Parse()

	utils.CheckRequiredFlag("instance", f.instance)
	utils.CheckRequiredFlag("platform-namespace", f.platformNamespace)
	utils.CheckRequiredFlag("component", f.compName)
	utils.CheckRequiredFlag("component-service-name", f.compSvcName)
	utils.CheckRequiredFlag("component-ip", f.compIP)
	utils.CheckRequiredFlag("vault-url", f.vaultURL)

	logkf.Global = logkf.
		BuildLoggerOrDie(f.logFormat, f.logLevel).
		WithPlatformComponent(api.PlatformComponentBootstrap)
	defer logkf.Global.Sync()
	log := logkf.Global

	log.DebugInterface("flags:", f)

	for retry := 0; retry < 3; retry++ {
		log.Infof("generating certificates for component %s, attempt %d of 3", f.compName, retry+1)
		if err := generateCerts(log); err != nil {
			log.Errorf("error generating certificates: %v", err)
			time.Sleep(time.Second * time.Duration(rand.Intn(2)+1))

		} else {
			return
		}
	}

	log.Fatal("unable to bootstrap component")
}

func generateCerts(log *logkf.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	key := vault.Key{
		Instance:  f.instance,
		Namespace: f.platformNamespace,
		Component: f.compName,
	}

	vaultCli, err := vault.New(vault.ClientOptions{
		Instance: f.instance,
		Role:     vault.RoleName(key),
		URL:      f.vaultURL,
		CACert:   api.PathCACert,
	})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%s", vault.PKISubPath(key, "issue"), vault.RoleName(key))
	s, err := vaultCli.Logical().WriteWithContext(ctx, path, map[string]interface{}{
		"common_name": fmt.Sprintf("%s.svc", f.compSvcName),
		"alt_names":   fmt.Sprintf("%s@%s,%s,localhost", f.compName, f.compSvcName, f.compSvcName),
		"ip_sans":     fmt.Sprintf("%s,127.0.0.1", f.compIP),
		"ttl":         TenYears,
	})
	if err != nil {
		return err
	}

	cert := fmt.Sprintf("%s\n%s", s.Data["certificate"], s.Data["issuing_ca"])
	privKey := fmt.Sprintf("%s", s.Data["private_key"])

	if err := os.WriteFile(api.PathTLSCert, []byte(cert), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls certificate to %s", api.PathTLSCert)

	if err := os.WriteFile(api.PathTLSKey, []byte(privKey), 0600); err != nil {
		return err
	}
	log.Infof("wrote tls private key to %s", api.PathTLSKey)

	return nil
}
