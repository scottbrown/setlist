package setlist

import (
	"fmt"
	"strings"

	"github.com/go-ini/ini"
)

// FileBuilder is responsible for generating an INI file
// based on the provided configuration.
type FileBuilder struct {
	Config ConfigFile
}

// NewFileBuilder creates a new FileBuilder instance with the
// given configuration.
func NewFileBuilder(configFile ConfigFile) FileBuilder {
	return FileBuilder{
		Config: configFile,
	}
}

// validateConfig checks if the configuration has all required fields set.
func (f *FileBuilder) validateConfig() error {
	if f.Config.SessionName == "" {
		return fmt.Errorf("missing required field: SessionName")
	}

	if f.Config.IdentityStoreId == "" && f.Config.FriendlyName == "" {
		return fmt.Errorf("missing required field: either IdentityStoreId or FriendlyName must be provided")
	}

	if f.Config.Region == "" {
		return fmt.Errorf("missing required field: Region")
	}

	return nil
}

// Build generates an INI file based on the configuration.
// It adds a default section, an SSO section, and profile sections
// for each configured profile. If a nickname mapping exists,
// it creates an additional profile section for the nickname.
func (f *FileBuilder) Build() (*ini.File, error) {
	// First validate the configuration
	if err := f.validateConfig(); err != nil {
		return nil, err
	}

	payload := ini.Empty()

	if err := f.addDefaultSection(payload); err != nil {
		return payload, err
	}

	if err := f.addSSOSection(payload); err != nil {
		return payload, err
	}

	for _, p := range f.Config.Profiles {
		p.Name = fmt.Sprintf("%s-%s", p.AccountId, p.RoleName)
		if err := f.addProfileSection(p, payload); err != nil {
			return payload, err
		}

		// Check if the profile has an associated nickname
		if !f.Config.HasNickname(p.AccountId) {
			continue
		}

		// Create section for AccountNickname-PermissionSet profile
		p.Name = fmt.Sprintf("%s-%s", f.Config.NicknameMapping[p.AccountId], p.RoleName)
		if err := f.addProfileSection(p, payload); err != nil {
			return payload, err
		}
	}

	return payload, nil
}

// addDefaultSection creates the [default] section in the INI file.
// This section contains the SSO session name and a timestamp comment.
func (f *FileBuilder) addDefaultSection(file *ini.File) error {
	section := file.Section("default")

	// Add a comment indicating when the file was generated
	section.Comment = fmt.Sprintf("# Generated on: %s", generateTimestamp())

	if _, err := section.NewKey(SSOSessionAttrKey, f.Config.SessionName); err != nil {
		return err
	}

	return nil
}

// addSSOSection creates the SSO session section in the INI file.
// This section includes the SSO start URL, region, and registration scopes.
func (f *FileBuilder) addSSOSection(file *ini.File) error {
	section := file.Section(strings.Join([]string{SSOSessionSectionKey, f.Config.SessionName}, " "))

	if _, err := section.NewKey(SSOStartUrlKey, f.Config.StartURL()); err != nil {
		return err
	}

	if _, err := section.NewKey(SSORegionKey, f.Config.Region); err != nil {
		return err
	}

	if _, err := section.NewKey(SSORegistrationScopesKey, SSORegistrationScopesValue); err != nil {
		return err
	}

	return nil
}

// addProfileSection creates a profile section in the INI file for a given
// profile. It includes metadata such as session name, account ID, and
// role name.
func (f *FileBuilder) addProfileSection(p Profile, file *ini.File) error {
	// Validate profile fields
	if p.SessionName == "" {
		return fmt.Errorf("profile missing required field: SessionName")
	}

	if p.AccountId == "" {
		return fmt.Errorf("profile missing required field: AccountId")
	}

	if p.RoleName == "" {
		return fmt.Errorf("profile missing required field: RoleName")
	}

	section := file.Section(fmt.Sprintf("profile %s", p.Name))

	// Add a comment describing the profile and session duration
	section.Comment = fmt.Sprintf("# %s. Session Duration: %s", p.Description, p.SessionDuration)

	if _, err := section.NewKey(SSOSessionAttrKey, p.SessionName); err != nil {
		return err
	}

	if _, err := section.NewKey(SSOAccountIdKey, p.AccountId); err != nil {
		return err
	}

	if _, err := section.NewKey(SSORoleNameKey, p.RoleName); err != nil {
		return err
	}

	return nil
}
