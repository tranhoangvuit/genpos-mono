package usecase

import (
	"context"

	"github.com/genpick/genpos-mono/backend/internal/domain/entity"
	"github.com/genpick/genpos-mono/backend/internal/usecase/input"
)

//go:generate mockgen -source=usecase.go -destination=mock/mock_usecase.go -package=mock

// StockTakeUsecase is the service contract consumed by the StockTakeService handler.
type StockTakeUsecase interface {
	ListStockTakes(ctx context.Context, orgID string) ([]*entity.StockTakeListItem, error)
	GetStockTake(ctx context.Context, in input.GetStockTakeInput) (*entity.StockTake, error)
	CreateStockTake(ctx context.Context, in input.CreateStockTakeInput) (*entity.StockTake, error)
	SaveStockTakeProgress(ctx context.Context, in input.SaveStockTakeProgressInput) (*entity.StockTake, error)
	FinalizeStockTake(ctx context.Context, in input.FinalizeStockTakeInput) (*entity.StockTake, error)
	CancelStockTake(ctx context.Context, in input.CancelStockTakeInput) (*entity.StockTake, error)
	DeleteStockTake(ctx context.Context, in input.DeleteStockTakeInput) error
}

// SupplierUsecase is the service contract consumed by the SupplierService handler.
type SupplierUsecase interface {
	ListSuppliers(ctx context.Context, orgID string) ([]*entity.Supplier, error)
	CreateSupplier(ctx context.Context, in input.CreateSupplierInput) (*entity.Supplier, error)
	UpdateSupplier(ctx context.Context, in input.UpdateSupplierInput) (*entity.Supplier, error)
	DeleteSupplier(ctx context.Context, in input.DeleteSupplierInput) error
}

// StoreUsecase is the service contract consumed by the StoreService handler.
type StoreUsecase interface {
	ListStoreDetails(ctx context.Context, orgID string) ([]*entity.Store, error)
	CreateStore(ctx context.Context, in input.CreateStoreInput) (*entity.Store, error)
	UpdateStore(ctx context.Context, in input.UpdateStoreInput) (*entity.Store, error)
	DeleteStore(ctx context.Context, in input.DeleteStoreInput) error
}

// PaymentMethodUsecase is the service contract consumed by the PaymentMethodService handler.
type PaymentMethodUsecase interface {
	ListPaymentMethods(ctx context.Context, orgID string) ([]*entity.PaymentMethod, error)
	CreatePaymentMethod(ctx context.Context, in input.CreatePaymentMethodInput) (*entity.PaymentMethod, error)
	UpdatePaymentMethod(ctx context.Context, in input.UpdatePaymentMethodInput) (*entity.PaymentMethod, error)
	DeletePaymentMethod(ctx context.Context, in input.DeletePaymentMethodInput) error
}

// TaxRateUsecase is the service contract consumed by the TaxRateService handler.
type TaxRateUsecase interface {
	ListTaxRates(ctx context.Context, orgID string) ([]*entity.TaxRate, error)
	CreateTaxRate(ctx context.Context, in input.CreateTaxRateInput) (*entity.TaxRate, error)
	UpdateTaxRate(ctx context.Context, in input.UpdateTaxRateInput) (*entity.TaxRate, error)
	DeleteTaxRate(ctx context.Context, in input.DeleteTaxRateInput) error
}

// MemberUsecase is the service contract consumed by the MemberService handler.
type MemberUsecase interface {
	ListMembers(ctx context.Context, orgID string) ([]*entity.Member, error)
	ListRoleOptions(ctx context.Context, orgID string) ([]*entity.RoleOption, error)
	CreateMember(ctx context.Context, in input.CreateMemberInput) (*entity.Member, error)
	UpdateMember(ctx context.Context, in input.UpdateMemberInput) (*entity.Member, error)
	DeleteMember(ctx context.Context, in input.DeleteMemberInput) error
}

// PurchaseOrderUsecase is the service contract consumed by the PurchaseOrderService handler.
type PurchaseOrderUsecase interface {
	ListPurchaseOrders(ctx context.Context, orgID string) ([]*entity.PurchaseOrderListItem, error)
	ListStores(ctx context.Context, orgID string) ([]*entity.StoreRef, error)
	ListVariantsForPicker(ctx context.Context, orgID string) ([]*entity.VariantPickerItem, error)
	GetPurchaseOrder(ctx context.Context, in input.GetPurchaseOrderInput) (*entity.PurchaseOrder, error)
	CreatePurchaseOrder(ctx context.Context, in input.CreatePurchaseOrderInput) (*entity.PurchaseOrder, error)
	UpdatePurchaseOrder(ctx context.Context, in input.UpdatePurchaseOrderInput) (*entity.PurchaseOrder, error)
	SubmitPurchaseOrder(ctx context.Context, in input.SubmitPurchaseOrderInput) (*entity.PurchaseOrder, error)
	ReceivePurchaseOrder(ctx context.Context, in input.ReceivePurchaseOrderInput) (*entity.PurchaseOrder, error)
	CancelPurchaseOrder(ctx context.Context, in input.CancelPurchaseOrderInput) (*entity.PurchaseOrder, error)
	DeletePurchaseOrder(ctx context.Context, in input.DeletePurchaseOrderInput) error
}

// OrderUsecase is the service contract consumed by the OrderService handler.
type OrderUsecase interface {
	ListOrders(ctx context.Context, in input.ListDailySalesInput) ([]*entity.OrderSummary, error)
	GetOrder(ctx context.Context, in input.GetOrderInput) (*entity.Order, error)
}

// CustomerUsecase is the service contract consumed by the CustomerService handler.
type CustomerUsecase interface {
	// Customers
	ListCustomers(ctx context.Context, orgID string) ([]*entity.CustomerListItem, error)
	GetCustomer(ctx context.Context, in input.GetCustomerInput) (*entity.Customer, error)
	CreateCustomer(ctx context.Context, in input.CreateCustomerInput) (*entity.Customer, error)
	UpdateCustomer(ctx context.Context, in input.UpdateCustomerInput) (*entity.Customer, error)
	DeleteCustomer(ctx context.Context, in input.DeleteCustomerInput) error

	// Customer groups
	ListCustomerGroups(ctx context.Context, orgID string) ([]*entity.CustomerGroup, error)
	CreateCustomerGroup(ctx context.Context, in input.CreateCustomerGroupInput) (*entity.CustomerGroup, error)
	UpdateCustomerGroup(ctx context.Context, in input.UpdateCustomerGroupInput) (*entity.CustomerGroup, error)
	DeleteCustomerGroup(ctx context.Context, in input.DeleteCustomerGroupInput) error

	// CSV import
	ParseImportCustomerCsv(ctx context.Context, in input.ParseImportCustomerCsvInput) (*input.ParseImportCustomerCsvResult, error)
	ImportCustomers(ctx context.Context, in input.ImportCustomersInput) (*input.ImportCustomersResult, error)
}

// CatalogUsecase is the service contract consumed by the CatalogService handler.
type CatalogUsecase interface {
	// Categories
	ListCategories(ctx context.Context, orgID string) ([]*entity.Category, error)
	CreateCategory(ctx context.Context, in input.CreateCategoryInput) (*entity.Category, error)
	UpdateCategory(ctx context.Context, in input.UpdateCategoryInput) (*entity.Category, error)
	DeleteCategory(ctx context.Context, in input.DeleteCategoryInput) error

	// Products
	ListProducts(ctx context.Context, orgID string) ([]*entity.ProductListItem, error)
	GetProduct(ctx context.Context, in input.GetProductInput) (*entity.ProductDetail, error)
	CreateProduct(ctx context.Context, in input.CreateProductInput) (*entity.ProductDetail, error)
	UpdateProduct(ctx context.Context, in input.UpdateProductInput) (*entity.ProductDetail, error)
	DeleteProduct(ctx context.Context, in input.DeleteProductInput) error

	// CSV import
	ParseImportCsv(ctx context.Context, in input.ParseImportCsvInput) (*input.ParseImportCsvResult, error)
	ImportProducts(ctx context.Context, in input.ImportProductsInput) (*input.ImportProductsResult, error)
}
