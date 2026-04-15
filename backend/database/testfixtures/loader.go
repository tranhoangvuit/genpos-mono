package testfixtures

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"
)

type Loader struct {
	db       *pgxpool.Pool
	basePath string
}

func NewLoader(db *pgxpool.Pool) *Loader {
	return &Loader{
		db: db,
	}
}

func NewLoaderWithPath(db *pgxpool.Pool, basePath string) *Loader {
	return &Loader{
		db:       db,
		basePath: basePath,
	}
}

func (l *Loader) LoadAll(ctx context.Context) error {
	tables := []string{
		"organizations",
		"stores",
		"registers",
		"roles",
		"users",
		"user_stores",
		"categories",
		"products",
		"product_variants",
		"tax_rates",
		"discounts",
		"customers",
		"payment_methods",
		"shifts",
		"orders",
		"order_line_items",
		"payments",
		"refunds",
		"refund_line_items",
		"stock_movements",
		"stock_cost_prices",
		"stock_cost_price_tracks",
		"purchase_orders",
		"purchase_order_items",
		"stock_takes",
		"stock_take_items",
		"store_config",
	}

	for _, table := range tables {
		if err := l.loadTable(ctx, table); err != nil {
			return fmt.Errorf("load table %s: %w", table, err)
		}
	}

	return nil
}

func (l *Loader) LoadTable(ctx context.Context, table string) error {
	return l.loadTable(ctx, table)
}

func (l *Loader) loadTable(ctx context.Context, table string) error {
	path := l.getTablePath(table)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	var rows []map[string]interface{}
	if err := yaml.Unmarshal(data, &rows); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	if len(rows) == 0 {
		return nil
	}

	columns := make([]string, 0, len(rows[0]))
	for col := range rows[0] {
		columns = append(columns, col)
	}

	if err := l.truncateTable(ctx, table); err != nil {
		return fmt.Errorf("truncate table: %w", err)
	}

	for _, row := range rows {
		if err := l.insertRow(ctx, table, columns, row); err != nil {
			return fmt.Errorf("insert row: %w", err)
		}
	}

	return nil
}

func (l *Loader) truncateTable(ctx context.Context, table string) error {
	_, err := l.db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	return err
}

func (l *Loader) insertRow(ctx context.Context, table string, columns []string, row map[string]interface{}) error {
	placeholders := make([]string, len(columns))
	values := make([]interface{}, len(columns))

	for i, col := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		values[i] = row[col]
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	_, err := l.db.Exec(ctx, query, values...)
	return err
}

func (l *Loader) getTablePath(table string) string {
	if l.basePath != "" {
		return filepath.Join(l.basePath, table+".yml")
	}
	return filepath.Join("database", "testfixtures", table+".yml")
}

func LoadAllFixtures(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadAll(ctx)
}

func LoadTableFixtures(ctx context.Context, db *pgxpool.Pool, table string) error {
	loader := NewLoader(db)
	return loader.LoadTable(ctx, table)
}

func LoadFixturesFromPath(ctx context.Context, db *pgxpool.Pool, basePath string) error {
	loader := NewLoaderWithPath(db, basePath)
	return loader.LoadAll(ctx)
}

type FixtureFile struct {
	Table  string
	Column string
}

func (l *Loader) LoadFixtureFiles(ctx context.Context, files []FixtureFile) error {
	for _, file := range files {
		if err := l.LoadTable(ctx, file.Table); err != nil {
			return fmt.Errorf("load %s: %w", file.Table, err)
		}
	}
	return nil
}

type FixtureByName struct {
	Org      string
	Store    string
	Register string
	Role     string
	User     string
	Category string
	Product  string
	Variant  string
	TaxRate  string
	Discount string
	Customer string
	Payment  string
	Order    string
	Shift    string
}

func (l *Loader) LoadByNames(ctx context.Context, names FixtureByName) error {
	loadOrder := []string{}

	if names.Org != "" || names.Store != "" || names.Register != "" {
		loadOrder = append(loadOrder, "organizations", "stores", "registers")
	}
	if names.Role != "" || names.User != "" {
		loadOrder = append(loadOrder, "roles", "users", "user_stores")
	}
	if names.Category != "" || names.Product != "" || names.Variant != "" || names.TaxRate != "" || names.Discount != "" {
		loadOrder = append(loadOrder, "categories", "products", "product_variants", "tax_rates", "discounts")
	}
	if names.Customer != "" || names.Payment != "" {
		loadOrder = append(loadOrder, "customers", "payment_methods")
	}
	if names.Order != "" || names.Shift != "" {
		loadOrder = append(loadOrder, "shifts", "orders", "order_line_items", "payments", "refunds", "refund_line_items")
	}

	for _, table := range loadOrder {
		if err := l.LoadTable(ctx, table); err != nil {
			return err
		}
	}

	return nil
}

func TableFile(table string) string {
	return table + ".yml"
}

var allTables = []string{
	"organizations",
	"stores",
	"registers",
	"roles",
	"users",
	"user_stores",
	"categories",
	"products",
	"product_variants",
	"tax_rates",
	"discounts",
	"customers",
	"payment_methods",
	"shifts",
	"orders",
	"order_line_items",
	"payments",
	"refunds",
	"refund_line_items",
	"stock_movements",
	"stock_cost_prices",
	"stock_cost_price_tracks",
	"purchase_orders",
	"purchase_order_items",
	"stock_takes",
	"stock_take_items",
	"store_config",
}

func AllTables() []string {
	return allTables
}

func FoundationTables() []string {
	return []string{"organizations", "stores", "registers"}
}

func AuthTables() []string {
	return []string{"roles", "users", "user_stores"}
}

func CatalogTables() []string {
	return []string{"categories", "products", "product_variants", "tax_rates", "discounts"}
}

func CustomerTables() []string {
	return []string{"customers", "payment_methods"}
}

func OrderTables() []string {
	return []string{"shifts", "orders", "order_line_items", "payments", "refunds", "refund_line_items"}
}

func InventoryTables() []string {
	return []string{"stock_movements", "stock_cost_prices", "stock_cost_price_tracks", "purchase_orders", "purchase_order_items", "stock_takes", "stock_take_items"}
}

func OperationTables() []string {
	return []string{"store_config"}
}

type Tier int

const (
	TierFoundation Tier = iota + 1
	TierAuth
	TierCatalog
	TierCustomer
	TierOperations
	TierOrders
	TierInventory
)

func (l *Loader) LoadTier(ctx context.Context, tier Tier) error {
	var tables []string
	switch tier {
	case TierFoundation:
		tables = FoundationTables()
	case TierAuth:
		tables = AuthTables()
	case TierCatalog:
		tables = CatalogTables()
	case TierCustomer:
		tables = CustomerTables()
	case TierOperations:
		tables = OperationTables()
	case TierOrders:
		tables = OrderTables()
	case TierInventory:
		tables = InventoryTables()
	default:
		return fmt.Errorf("unknown tier: %d", tier)
	}

	for _, table := range tables {
		if err := l.LoadTable(ctx, table); err != nil {
			return err
		}
	}
	return nil
}

func LoadFoundation(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierFoundation)
}

func LoadAuth(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierAuth)
}

func LoadCatalog(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierCatalog)
}

func LoadCustomer(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierCustomer)
}

func LoadOrders(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierOrders)
}

func LoadInventory(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierInventory)
}

func LoadOperations(ctx context.Context, db *pgxpool.Pool) error {
	loader := NewLoader(db)
	return loader.LoadTier(ctx, TierOperations)
}

type FixtureLoader interface {
	LoadAllFixtures(ctx context.Context, db *pgxpool.Pool) error
}

type TieredLoader struct {
	db     *pgxpool.Pool
	loaded map[Tier]bool
	loader *Loader
}

func NewTieredLoader(db *pgxpool.Pool) *TieredLoader {
	return &TieredLoader{
		db:     db,
		loaded: make(map[Tier]bool),
		loader: NewLoader(db),
	}
}

func (tl *TieredLoader) LoadTiers(ctx context.Context, tiers ...Tier) error {
	for _, tier := range tiers {
		if tl.loaded[tier] {
			continue
		}
		if err := tl.loader.LoadTier(ctx, tier); err != nil {
			return err
		}
		tl.loaded[tier] = true
	}
	return nil
}

func (tl *TieredLoader) LoadFoundation(ctx context.Context) error {
	return tl.LoadTiers(ctx, TierFoundation)
}

func (tl *TieredLoader) LoadAuth(ctx context.Context) error {
	if err := tl.LoadTiers(ctx, TierFoundation); err != nil {
		return err
	}
	return tl.LoadTiers(ctx, TierAuth)
}

func (tl *TieredLoader) LoadCatalog(ctx context.Context) error {
	if err := tl.LoadTiers(ctx, TierFoundation, TierAuth); err != nil {
		return err
	}
	return tl.LoadTiers(ctx, TierCatalog)
}

func (tl *TieredLoader) LoadCustomer(ctx context.Context) error {
	if err := tl.LoadTiers(ctx, TierFoundation, TierAuth); err != nil {
		return err
	}
	return tl.LoadTiers(ctx, TierCustomer)
}

func (tl *TieredLoader) LoadOrdersWithDeps(ctx context.Context) error {
	return tl.LoadTiers(ctx, TierFoundation, TierAuth, TierCatalog, TierCustomer, TierOperations, TierOrders)
}

func (tl *TieredLoader) LoadInventoryWithDeps(ctx context.Context) error {
	return tl.LoadTiers(ctx, TierFoundation, TierAuth, TierCatalog, TierCustomer, TierOperations, TierOrders, TierInventory)
}

func (tl *TieredLoader) LoadAll(ctx context.Context) error {
	return tl.LoadTiers(ctx, TierFoundation, TierAuth, TierCatalog, TierCustomer, TierOperations, TierOrders, TierInventory)
}

func (tl *TieredLoader) LoadFixturesForOrders(ctx context.Context) error {
	return tl.LoadOrdersWithDeps(ctx)
}

func (tl *TieredLoader) LoadFixturesForInventory(ctx context.Context) error {
	return tl.LoadInventoryWithDeps(ctx)
}
