// Command iam-bootstrap grants a platform-level IAM role (platform_owner or
// platform_admin) to an existing auth user, directly against the IAM
// database. It is the only supported way to create the first platform
// admin on a fresh deployment, and doubles as the break-glass recovery path
// when every existing platform admin is unreachable. See
// docs/09-pzep/PZEP-0002-platform-admin-bootstrap-and-recovery.md.
//
// This is intentionally a standalone, one-shot CLI with no HTTP/gRPC
// listener of its own - possession of IAM database credentials (via the
// same CONFIG_PATH the iam service itself uses) is the trust boundary, not
// a login. Run it from a trusted shell with database network access, the
// same trust level already required to run goose migrations:
//
//	CONFIG_PATH=... go run ./cmd/iam-bootstrap \
//	  --user-id 4 --role platform_owner --reason "initial platform admin for staging"
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/fx"

	"github.com/tuannm99/podzone/internal/iam"
	"github.com/tuannm99/podzone/internal/iam/server"
	"github.com/tuannm99/podzone/pkg/pdconfig"
	"github.com/tuannm99/podzone/pkg/pdlog"
	"github.com/tuannm99/podzone/pkg/pdsql"
)

func main() {
	var userID uint64
	var opts bootstrapOptions
	flag.Uint64Var(&userID, "user-id", 0, "target auth user ID (required)")
	flag.StringVar(&opts.Role, "role", "", "platform_owner or platform_admin (required)")
	flag.StringVar(&opts.Reason, "reason", "", "reason for this grant, stored in the audit log (required)")
	flag.BoolVar(&opts.Force, "force", false,
		"required to grant when the target role already has an active admin (break-glass recovery)")
	flag.Parse()
	opts.UserID = uint(userID)

	if err := validateFlags(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_ = godotenv.Load()

	var deps bootstrapDeps
	app := fx.New(
		pdconfig.Module,
		pdlog.Module,
		pdsql.ModuleFor("iam"),
		iam.CommandRepositoryModule,
		server.SharedModule,
		fx.Populate(&deps.Roles, &deps.MembershipCommands, &deps.MembershipQueries, &deps.Audit, &deps.Users),
		fx.NopLogger,
	)

	startCtx, cancelStart := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelStart()
	if err := app.Start(startCtx); err != nil {
		fmt.Fprintln(os.Stderr, "iam-bootstrap: startup failed:", err)
		os.Exit(1)
	}
	defer func() {
		stopCtx, cancelStop := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancelStop()
		_ = app.Stop(stopCtx)
	}()

	auditLogID, err := runBootstrap(context.Background(), deps, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("granted %s to user %d - audit_log_id=%s\n", opts.Role, opts.UserID, auditLogID)
}
