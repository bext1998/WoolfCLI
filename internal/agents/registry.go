package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Registry struct {
	roles   map[string]Role
	presets map[string]Preset
}

func NewRegistry(agentsDir string) (*Registry, error) {
	registry := &Registry{
		roles:   map[string]Role{},
		presets: map[string]Preset{},
	}
	for _, role := range BuiltinRoles() {
		if err := registry.AddRole(role); err != nil {
			return nil, err
		}
	}
	for _, preset := range BuiltinPresets() {
		if err := registry.AddPreset(preset); err != nil {
			return nil, err
		}
	}
	if agentsDir != "" {
		if err := registry.loadUserRoles(agentsDir); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) AddRole(role Role) error {
	if err := role.Validate(); err != nil {
		return err
	}
	r.roles[role.Name] = role
	return nil
}

func (r *Registry) AddPreset(preset Preset) error {
	if strings.TrimSpace(preset.Name) == "" {
		return fmt.Errorf("CFG-003: preset name is required")
	}
	if len(preset.Roles) == 0 {
		return fmt.Errorf("CFG-003: preset %s requires at least one role", preset.Name)
	}
	for _, name := range preset.Roles {
		if _, ok := r.roles[name]; !ok {
			return fmt.Errorf("CFG-003: preset %s references unknown role %s", preset.Name, name)
		}
	}
	r.presets[preset.Name] = preset
	return nil
}

func (r *Registry) Role(name string) (Role, bool) {
	role, ok := r.roles[name]
	return role, ok
}

func (r *Registry) ListRoles() []Role {
	items := make([]Role, 0, len(r.roles))
	for _, role := range r.roles {
		items = append(items, role)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func (r *Registry) Preset(name string) (Preset, bool) {
	preset, ok := r.presets[name]
	return preset, ok
}

func (r *Registry) ListPresets() []Preset {
	items := make([]Preset, 0, len(r.presets))
	for _, preset := range r.presets {
		items = append(items, preset)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	return items
}

func (r *Registry) ResolvePreset(name string) ([]Role, error) {
	preset, ok := r.Preset(name)
	if !ok {
		return nil, fmt.Errorf("CFG-003: preset %s not found", name)
	}
	return r.ResolveRoles(preset.Roles)
}

func (r *Registry) ResolveRoles(names []string) ([]Role, error) {
	roles := make([]Role, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		role, ok := r.Role(name)
		if !ok {
			return nil, fmt.Errorf("CFG-003: role %s not found", name)
		}
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("CFG-003: at least one role is required")
	}
	return roles, nil
}

func (r *Registry) loadUserRoles(agentsDir string) error {
	paths, err := filepath.Glob(filepath.Join(agentsDir, "*.yaml"))
	if err != nil {
		return err
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		role, err := LoadRole(path)
		if err != nil {
			return err
		}
		if err := r.AddRole(role); err != nil {
			return err
		}
	}
	return nil
}
