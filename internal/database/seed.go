package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/Snowitty-Re/e-fiber-admin/internal/config"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/permission"
	"github.com/Snowitty-Re/e-fiber-admin/internal/ent/role"
)

type permEntry struct {
	resource string
	actions  []string
}

var permCatalog = []permEntry{
	{"product", []string{"read", "write", "delete", "publish", "archive"}},
	{"variant", []string{"read", "write"}},
	{"category", []string{"read", "write", "delete"}},
	{"collection", []string{"read", "write", "delete"}},
	{"order", []string{"read", "write", "approve", "cancel", "fulfill", "refund", "export"}},
	{"customer", []string{"read", "write", "delete"}},
	{"inquiry", []string{"read", "assign", "update", "convert", "export"}},
	{"cms", []string{"read", "write", "delete", "publish"}},
	{"settings", []string{"read", "write", "dangerous"}},
	{"media", []string{"read", "write", "delete"}},
	{"region", []string{"read", "write"}},
	{"discount", []string{"read", "write", "delete"}},
	{"payment", []string{"read", "write"}},
	{"shipping", []string{"read", "write"}},
	{"admin", []string{"read", "write", "delete"}},
	{"webhook", []string{"read", "write", "delete"}},
	{"notification", []string{"read", "write"}},
}

type roleDef struct {
	slug        string
	name        string
	description string
	permFilter  func(resource, action string) bool
}

var roleDefs = []roleDef{
	{"owner", "Owner", "Full access (system)", func(r, a string) bool { return true }},
	{"admin", "Admin", "Full access except admin management & dangerous settings",
		func(r, a string) bool {
			if r == "admin" {
				return false
			}
			if r == "settings" && a == "dangerous" {
				return false
			}
			return true
		}},
	{"operator", "Operator", "Operations: products, orders, customers, inquiries",
		func(r, a string) bool {
			switch r {
			case "product", "variant", "category", "collection", "order", "customer",
				"inquiry", "cms", "media", "discount":
				return a == "read" || a == "write" || a == "publish" || a == "approve" ||
					a == "cancel" || a == "fulfill" || a == "refund" || a == "assign" ||
					a == "update" || a == "convert" || a == "delete"
			case "region", "payment", "shipping", "webhook", "notification":
				return a == "read"
			}
			return false
		}},
	{"content", "Content Editor", "CMS content editing",
		func(r, a string) bool {
			switch r {
			case "cms":
				return true
			case "product", "variant", "category", "collection":
				return a == "read"
			case "media":
				return a == "read" || a == "write"
			case "region":
				return a == "read"
			}
			return false
		}},
	{"support", "Support", "Customer service: read/update orders, inquiries, customers",
		func(r, a string) bool {
			switch r {
			case "order", "customer":
				return a == "read" || a == "write"
			case "inquiry":
				return a == "read" || a == "write" || a == "assign" || a == "update" || a == "convert"
			case "product", "variant":
				return a == "read"
			case "notification":
				return a == "read"
			}
			return false
		}},
	{"viewer", "Viewer", "Read-only access",
		func(r, a string) bool { return a == "read" }},
}

func Seed(ctx context.Context, client *ent.Client, cfg config.SeedConfig) error {
	perms, err := seedPermissions(ctx, client)
	if err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}
	slog.Info("permissions seeded", "count", len(perms))

	roles, err := seedRoles(ctx, client, perms)
	if err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}
	slog.Info("roles seeded", "count", len(roles))

	if err := seedOwner(ctx, client, cfg, roles["owner"]); err != nil {
		return fmt.Errorf("seed owner: %w", err)
	}
	slog.Info("owner ensured", "email", cfg.OwnerEmail)

	return nil
}

func seedPermissions(ctx context.Context, client *ent.Client) (map[string]*ent.Permission, error) {
	perms := make(map[string]*ent.Permission)
	for _, pe := range permCatalog {
		for _, action := range pe.actions {
			key := permKey(pe.resource, action)
			existing, err := client.Permission.Query().
				Where(permission.ResourceEQ(pe.resource), permission.ActionEQ(action)).
				Only(ctx)
			if err != nil && !ent.IsNotFound(err) {
				return nil, err
			}
			if existing != nil {
				perms[key] = existing
				continue
			}
			p, err := client.Permission.Create().
				SetResource(pe.resource).
				SetAction(action).
				SetDescription(fmt.Sprintf("%s:%s", pe.resource, action)).
				Save(ctx)
			if err != nil {
				return nil, err
			}
			perms[key] = p
		}
	}
	return perms, nil
}

func seedRoles(ctx context.Context, client *ent.Client, perms map[string]*ent.Permission) (map[string]*ent.Role, error) {
	roles := make(map[string]*ent.Role)
	for _, rd := range roleDefs {
		existing, err := client.Role.Query().Where(role.SlugEQ(rd.slug)).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, err
		}
		var r *ent.Role
		if existing != nil {
			r = existing
		} else {
			r, err = client.Role.Create().
				SetName(rd.name).
				SetSlug(rd.slug).
				SetDescription(rd.description).
				SetIsSystem(true).
				Save(ctx)
			if err != nil {
				return nil, err
			}
		}
		rolePerms := filterPerms(perms, rd.permFilter)
		if err := client.Role.UpdateOne(r).ClearPermissions().AddPermissionIDs(rolePerms...).Exec(ctx); err != nil {
			return nil, err
		}
		roles[rd.slug] = r
	}
	return roles, nil
}

func seedOwner(ctx context.Context, client *ent.Client, cfg config.SeedConfig, ownerRole *ent.Role) error {
	exists, err := client.AdminUser.Query().Where().Count(ctx)
	if err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.OwnerPassword), 12)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	parts := strings.SplitN(cfg.OwnerEmail, "@", 2)
	name := parts[0]
	_, err = client.AdminUser.Create().
		SetEmail(cfg.OwnerEmail).
		SetPasswordHash(string(hash)).
		SetFirstName(name).
		SetLastName("Owner").
		AddRoles(ownerRole).
		Save(ctx)
	return err
}

func filterPerms(perms map[string]*ent.Permission, filter func(string, string) bool) []int {
	var ids []int
	for key, p := range perms {
		if filter(p.Resource, p.Action) {
			ids = append(ids, p.ID)
		}
		_ = key
	}
	return ids
}

func permKey(resource, action string) string {
	return resource + ":" + action
}
