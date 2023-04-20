package cmd

import (
	// _ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"

	// cp "github.com/otiai10/copy"
	"sigs.k8s.io/yaml"
)

var opensearchCmd = &cobra.Command{
	Use:   "opensearch",
	Short: "",
	Long:  "",
	Run:   setupOpensearch,
}

var openSearchConfigPath = "/kubefox/opensearch/"
var opensearchSecurityDir = path.Join(openSearchConfigPath, "opensearch-security")

// go:embed opensearch-security-config.yml
// var securityConfig []byte

// go:embed opensearch-security-roles.yml
// var securityRoles []byte

func init() {
	bootstrapCmd.AddCommand(opensearchCmd)
	// bootstrapCmd.Flags().StringVarP()
}

// https://opensearch.org/docs/latest/security-plugin/configuration/yaml
func setupOpensearch(cmd *cobra.Command, args []string) {
	opensearchCertsDir := path.Join(openSearchConfigPath, "certs")
	pemCertFilePath := path.Join(opensearchCertsDir, "kf_opensearch.pem")
	pemKeyFilePath := path.Join(opensearchCertsDir, "kf_opensearch-key.pem")
	pemTrustedCasFilePath := path.Join(opensearchCertsDir, "kf_opensearch_root-ca.pem")
	jsonFileFromVault := "/tls-certs/certs.json"
	var certs map[string]interface{}
	jsonContents, err := os.ReadFile(jsonFileFromVault)
	if err != nil {
		log.Fatalf("error reading from %s: %v", jsonFileFromVault, err)
		return
	}
	if err = json.Unmarshal(jsonContents, &certs); err != nil {
		log.Fatalf("error unmarshaling json: %v", err)
	}
	certsData := certs["data"].(map[string]interface{})
	if err = os.MkdirAll(opensearchCertsDir, os.ModePerm); err != nil {
		log.Fatalf("error creating %s: %v", opensearchCertsDir, err)
	}
	os.WriteFile(pemCertFilePath, []byte((certsData["certificate"].(string))+"\n"), os.ModePerm)
	os.WriteFile(pemKeyFilePath, []byte((certsData["private_key"].(string))+"\n"), os.ModePerm)
	os.WriteFile(pemTrustedCasFilePath, []byte((certsData["issuing_ca"].(string))+"\n"), os.ModePerm)
	if err = generateInternalUsers(); err != nil {
		log.Fatalf("error generating internal_users.yml: %v", err)
	}

	// cp.Copy("/opensearch-security", opensearchSecurityDir, cp.Options{AddPermission: 0700})
}

type InternalUsers struct {
	Meta  Meta            `json:"_meta" yaml:"_meta"`
	Users map[string]User `json:",inline" yaml:",inline"`
}

type Meta struct {
	Type          string `json:"type"  yaml:"type"`
	ConfigVersion int    `json:"config_version"  yaml:"config_version"`
}

type User struct {
	Hash         string   `json:"hash"  yaml:"hash"`
	Reserved     bool     `json:"reserved"  yaml:"reserved"`
	BackendRoles []string `json:"backend_roles"  yaml:"backend_roles"`
	Description  string   `json:"description"  yaml:"description"`
}

// https://opensearch.org/docs/latest/security-plugin/configuration/yaml/#internal_usersyml
func generateInternalUsers() error {
	internalUsersPath := path.Join(opensearchSecurityDir, "internal_users.yml")
	passwordsFileFromVault := "/secrets/passwords.json"
	jsonContents, err := os.ReadFile(passwordsFileFromVault)
	if err != nil {
		return fmt.Errorf("error reading from %s: %v", passwordsFileFromVault, err)
	}
	var passwords map[string]string
	if err = json.Unmarshal(jsonContents, &passwords); err != nil {
		return fmt.Errorf("error unmarshaling json: %v", err)
	}
	adminPasswordHash, err := bcrypt.GenerateFromPassword([]byte(passwords["admin_password"]), 12)
	if err != nil {
		return fmt.Errorf("Error creating password hash: %v", err)
	}
	admin := User{
		Hash:         string(adminPasswordHash),
		Reserved:     true,
		BackendRoles: []string{"admin"},
		Description:  "Admin user",
	}
	internalUsers := InternalUsers{
		Meta: Meta{
			Type:          "internalusers",
			ConfigVersion: 2,
		},
		Users: map[string]User{"admin": admin},
	}
	internalUsersContent, err := yaml.Marshal(&internalUsers)
	if err != nil {
		return fmt.Errorf("Error marshaling into yaml: %v", err)
	}
	if err = os.MkdirAll(opensearchSecurityDir, os.ModePerm); err != nil {
		return fmt.Errorf("Error creating %s: %v", opensearchSecurityDir, err)
	}
	if err := os.WriteFile(internalUsersPath, internalUsersContent, os.ModePerm); err != nil {
		return fmt.Errorf("Error writing to internal_users.yml: %v", err)
	}
	return nil
}
