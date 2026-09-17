package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

func main() {
	apply := flag.Bool("apply", false, "apply the migration after a successful dry-run")
	rate := flag.String("rate", "", "expected legacy USD/CNY rate; required with --apply")
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: billing-cny-migrate [--apply --rate <captured legacy rate>]")
		fmt.Fprintln(flag.CommandLine.Output(), "Offline one-time CNY ledger migration. Stop every ledger writer and back up both databases first.")
		fmt.Fprintln(flag.CommandLine.Output(), "Database configuration: SQL_DSN / SQLITE_PATH and optional LOG_SQL_DSN.")
		fmt.Fprintln(flag.CommandLine.Output(), "Without --apply: validate and roll back all data changes. Exit 2 means migration is required; exit 1 means an error.")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *common.PrintHelp {
		flag.Usage()
		return
	}
	if flag.NArg() != 0 {
		log.Fatal("unexpected positional arguments; use --help for migration options")
	}
	if *apply && strings.TrimSpace(*rate) == "" {
		log.Fatal("--rate is required with --apply; run without --apply first and use its captured legacy rate")
	}
	common.InitEnv()
	if err := model.InitBillingCurrencyMigrationDatabases(); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := model.CloseDB(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	report, err := model.MigrateLegacyUSDLedgerToCNY(*rate, *apply)
	if err != nil {
		log.Fatal(err)
	}
	encoded, err := common.Marshal(report)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(encoded))
	if !*apply && !report.MainAlreadyMigrated {
		os.Exit(2)
	}
}
