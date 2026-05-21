package inputport

import routingusecase "github.com/tuannm99/podzone/internal/backoffice/domain/routing/usecase"

type (
	CreateRoutedOrderCmd             = routingusecase.CreateRoutedOrderCmd
	RecommendRoutedOrderPartnerQuery = routingusecase.RecommendRoutedOrderPartnerQuery
	OpenOrderExceptionCmd            = routingusecase.OpenOrderExceptionCmd
	UpdateOrderExceptionStatusCmd    = routingusecase.UpdateOrderExceptionStatusCmd
	UpdateOrderShipmentCmd           = routingusecase.UpdateOrderShipmentCmd
	UpdateOrderSettlementCmd         = routingusecase.UpdateOrderSettlementCmd
	UpdateOrderIssueHandlingCmd      = routingusecase.UpdateOrderIssueHandlingCmd
	UpdateOrderQueueControlCmd       = routingusecase.UpdateOrderQueueControlCmd
	BulkUpdateRoutedOrdersCmd        = routingusecase.BulkUpdateRoutedOrdersCmd
	ListRoutedOrderActivitiesQuery   = routingusecase.ListRoutedOrderActivitiesQuery
	OrderRoutingUsecase              = routingusecase.OrderRoutingUsecase
)
