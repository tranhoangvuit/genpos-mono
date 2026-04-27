package interceptor

import (
	"context"

	"connectrpc.com/connect"

	"github.com/genpick/genpos-mono/backend/pkg/auth"
	"github.com/genpick/genpos-mono/backend/pkg/errors"
)

// ProcedurePermission maps a procedure name to its required resource and action.
type ProcedurePermission struct {
	Resource string
	Action   string
}

// NewPermissionInterceptor enforces that the caller's JWT permissions grant
// the resource:action required by the procedure. Public and auth-only
// procedures must be excluded from the map — they are handled by the auth
// interceptor or explicitly skipped here.
func NewPermissionInterceptor(rules map[string]ProcedurePermission) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			proc := req.Spec().Procedure

			if IsPublicProcedure(proc) {
				return next(ctx, req)
			}

			rule, mapped := rules[proc]
			if !mapped {
				return next(ctx, req)
			}

			authCtx := FromContext(ctx)
			if authCtx == nil {
				return nil, errors.Unauthorized("not signed in")
			}

			perms := auth.PermissionSet(authCtx.Permissions)
			if !perms.Allows(rule.Resource, rule.Action) {
				return nil, errors.Forbidden("insufficient permissions")
			}

			return next(ctx, req)
		})
	}
}

// DefaultProcedurePermissions returns the initial procedure → permission map.
func DefaultProcedurePermissions() map[string]ProcedurePermission {
	return map[string]ProcedurePermission{
		// Catalog — categories
		"/genpos.v1.CatalogService/ListCategories":  {Resource: "products", Action: "read"},
		"/genpos.v1.CatalogService/CreateCategory":  {Resource: "products", Action: "write"},
		"/genpos.v1.CatalogService/UpdateCategory":  {Resource: "products", Action: "write"},
		"/genpos.v1.CatalogService/DeleteCategory":  {Resource: "products", Action: "write"},

		// Catalog — products
		"/genpos.v1.CatalogService/ListProducts":   {Resource: "products", Action: "read"},
		"/genpos.v1.CatalogService/GetProduct":     {Resource: "products", Action: "read"},
		"/genpos.v1.CatalogService/CreateProduct":  {Resource: "products", Action: "write"},
		"/genpos.v1.CatalogService/UpdateProduct":  {Resource: "products", Action: "write"},
		"/genpos.v1.CatalogService/DeleteProduct":  {Resource: "products", Action: "write"},

		// Catalog — CSV import
		"/genpos.v1.CatalogService/ParseImportCsv": {Resource: "products", Action: "write"},
		"/genpos.v1.CatalogService/ImportProducts": {Resource: "products", Action: "write"},

		// Customers
		"/genpos.v1.CustomerService/ListCustomers":          {Resource: "customers", Action: "read"},
		"/genpos.v1.CustomerService/GetCustomer":            {Resource: "customers", Action: "read"},
		"/genpos.v1.CustomerService/ListCustomerGroups":     {Resource: "customers", Action: "read"},
		"/genpos.v1.CustomerService/CreateCustomer":         {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/UpdateCustomer":         {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/DeleteCustomer":         {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/CreateCustomerGroup":    {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/UpdateCustomerGroup":    {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/DeleteCustomerGroup":    {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/ParseImportCustomerCsv": {Resource: "customers", Action: "write"},
		"/genpos.v1.CustomerService/ImportCustomers":        {Resource: "customers", Action: "write"},

		// Suppliers
		"/genpos.v1.SupplierService/ListSuppliers":  {Resource: "inventory", Action: "read"},
		"/genpos.v1.SupplierService/CreateSupplier": {Resource: "inventory", Action: "write"},
		"/genpos.v1.SupplierService/UpdateSupplier": {Resource: "inventory", Action: "write"},
		"/genpos.v1.SupplierService/DeleteSupplier": {Resource: "inventory", Action: "write"},

		// Purchase orders
		"/genpos.v1.PurchaseOrderService/ListPurchaseOrders":   {Resource: "inventory", Action: "read"},
		"/genpos.v1.PurchaseOrderService/ListStores":           {Resource: "inventory", Action: "read"},
		"/genpos.v1.PurchaseOrderService/ListVariantsForPicker": {Resource: "inventory", Action: "read"},
		"/genpos.v1.PurchaseOrderService/GetPurchaseOrder":     {Resource: "inventory", Action: "read"},
		"/genpos.v1.PurchaseOrderService/CreatePurchaseOrder":  {Resource: "inventory", Action: "write"},
		"/genpos.v1.PurchaseOrderService/UpdatePurchaseOrder":  {Resource: "inventory", Action: "write"},
		"/genpos.v1.PurchaseOrderService/SubmitPurchaseOrder":  {Resource: "inventory", Action: "write"},
		"/genpos.v1.PurchaseOrderService/ReceivePurchaseOrder": {Resource: "inventory", Action: "write"},
		"/genpos.v1.PurchaseOrderService/CancelPurchaseOrder":  {Resource: "inventory", Action: "write"},
		"/genpos.v1.PurchaseOrderService/DeletePurchaseOrder":  {Resource: "inventory", Action: "write"},

		// Stores (settings)
		"/genpos.v1.StoreService/ListStoreDetails": {Resource: "settings", Action: "read"},
		"/genpos.v1.StoreService/CreateStore":      {Resource: "settings", Action: "write"},
		"/genpos.v1.StoreService/UpdateStore":      {Resource: "settings", Action: "write"},
		"/genpos.v1.StoreService/DeleteStore":      {Resource: "settings", Action: "write"},

		// Payment methods (settings)
		"/genpos.v1.PaymentMethodService/ListPaymentMethods":  {Resource: "settings", Action: "read"},
		"/genpos.v1.PaymentMethodService/CreatePaymentMethod": {Resource: "settings", Action: "write"},
		"/genpos.v1.PaymentMethodService/UpdatePaymentMethod": {Resource: "settings", Action: "write"},
		"/genpos.v1.PaymentMethodService/DeletePaymentMethod": {Resource: "settings", Action: "write"},

		// Tax rates (settings)
		"/genpos.v1.TaxRateService/ListTaxRates":  {Resource: "settings", Action: "read"},
		"/genpos.v1.TaxRateService/CreateTaxRate": {Resource: "settings", Action: "write"},
		"/genpos.v1.TaxRateService/UpdateTaxRate": {Resource: "settings", Action: "write"},
		"/genpos.v1.TaxRateService/DeleteTaxRate": {Resource: "settings", Action: "write"},

		// Members (users)
		"/genpos.v1.MemberService/ListMembers":  {Resource: "users", Action: "read"},
		"/genpos.v1.MemberService/ListRoles":    {Resource: "users", Action: "read"},
		"/genpos.v1.MemberService/CreateMember": {Resource: "users", Action: "write"},
		"/genpos.v1.MemberService/UpdateMember": {Resource: "users", Action: "write"},
		"/genpos.v1.MemberService/DeleteMember": {Resource: "users", Action: "write"},

		// Orders (reports)
		"/genpos.v1.OrderService/ListOrders":  {Resource: "orders", Action: "read"},
		"/genpos.v1.OrderService/GetOrder":    {Resource: "orders", Action: "read"},
		"/genpos.v1.OrderService/CreateOrder": {Resource: "orders", Action: "write"},

		// Stock takes
		"/genpos.v1.StockTakeService/ListStockTakes":        {Resource: "inventory", Action: "read"},
		"/genpos.v1.StockTakeService/GetStockTake":          {Resource: "inventory", Action: "read"},
		"/genpos.v1.StockTakeService/CreateStockTake":       {Resource: "inventory", Action: "write"},
		"/genpos.v1.StockTakeService/SaveStockTakeProgress": {Resource: "inventory", Action: "write"},
		"/genpos.v1.StockTakeService/FinalizeStockTake":     {Resource: "inventory", Action: "write"},
		"/genpos.v1.StockTakeService/CancelStockTake":       {Resource: "inventory", Action: "write"},
		"/genpos.v1.StockTakeService/DeleteStockTake":       {Resource: "inventory", Action: "write"},
	}
}
