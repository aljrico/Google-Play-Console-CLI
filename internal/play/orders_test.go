package play

import (
	"context"
	"reflect"
	"testing"
)

func TestGetOrderPassesOptionsToGetter(t *testing.T) {
	getter := &fakeOrderGetter{result: OrderGetResult{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
		Order:       Order{OrderID: "GPA.123", LineItems: []OrderLineItem{}},
	}}
	options := OrderGetOptions{PackageName: "com.example.app", OrderID: "GPA.123"}

	result, err := GetOrder(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("GetOrder() error = %v", err)
	}
	if result.Order.OrderID != "GPA.123" {
		t.Fatalf("OrderID = %q, want GPA.123", result.Order.OrderID)
	}
	if !reflect.DeepEqual(getter.options, options) {
		t.Fatalf("options = %#v, want %#v", getter.options, options)
	}
}

func TestGetOrderRejectsInvalidOptions(t *testing.T) {
	tests := []OrderGetOptions{
		{},
		{PackageName: "bad", OrderID: "GPA.123"},
		{PackageName: "com.example.app"},
	}
	for _, options := range tests {
		_, err := GetOrder(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("GetOrder(%#v) expected validation error", options)
		}
	}
}

func TestBatchGetOrdersPassesOptionsToGetter(t *testing.T) {
	getter := &fakeOrderBatchGetter{result: OrderBatchGetResult{
		PackageName: "com.example.app",
		OrderIDs:    []OrderID{"GPA.123", "GPA.456"},
		Orders:      []Order{{OrderID: "GPA.123", LineItems: []OrderLineItem{}}},
	}}
	options := OrderBatchGetOptions{
		PackageName: "com.example.app",
		OrderIDs:    []OrderID{"GPA.123", "GPA.456"},
	}

	result, err := BatchGetOrders(context.Background(), getter, options)
	if err != nil {
		t.Fatalf("BatchGetOrders() error = %v", err)
	}
	if len(result.Orders) != 1 {
		t.Fatalf("len(Orders) = %d, want 1", len(result.Orders))
	}
	if !reflect.DeepEqual(getter.options, options) {
		t.Fatalf("options = %#v, want %#v", getter.options, options)
	}
}

func TestBatchGetOrdersRejectsInvalidOptions(t *testing.T) {
	tests := []OrderBatchGetOptions{
		{},
		{PackageName: "com.example.app"},
		{PackageName: "com.example.app", OrderIDs: []OrderID{""}},
		{PackageName: "com.example.app", OrderIDs: []OrderID{"GPA.123", "GPA.123"}},
	}
	for _, options := range tests {
		_, err := BatchGetOrders(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("BatchGetOrders(%#v) expected validation error", options)
		}
	}
}

func TestRefundOrderDryRunDoesNotRequireRefunder(t *testing.T) {
	result, err := RefundOrder(context.Background(), nil, OrderRefundOptions{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
		Revoke:      true,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("RefundOrder() error = %v", err)
	}
	if !result.DryRun || result.Applied {
		t.Fatalf("result = %#v, want dry-run not applied", result)
	}
	if !result.Plan.Revoke || len(result.Plan.Steps) != 2 {
		t.Fatalf("plan = %#v, want refund and revoke steps", result.Plan)
	}
}

func TestRefundOrderAppliesWithConfirm(t *testing.T) {
	refunder := &fakeOrderRefunder{}
	result, err := RefundOrder(context.Background(), refunder, OrderRefundOptions{
		PackageName: "com.example.app",
		OrderID:     "GPA.123",
		Confirm:     true,
	})
	if err != nil {
		t.Fatalf("RefundOrder() error = %v", err)
	}
	if !result.Applied || refunder.options.OrderID != "GPA.123" {
		t.Fatalf("result = %#v options = %#v, want applied refund", result, refunder.options)
	}
}

func TestRefundOrderRejectsInvalidOptions(t *testing.T) {
	tests := []OrderRefundOptions{
		{},
		{PackageName: "bad", OrderID: "GPA.123", DryRun: true},
		{PackageName: "com.example.app", DryRun: true},
		{PackageName: "com.example.app", OrderID: "GPA.123"},
	}
	for _, options := range tests {
		_, err := RefundOrder(context.Background(), nil, options)
		if err == nil {
			t.Fatalf("RefundOrder(%#v) expected validation error", options)
		}
	}
}

type fakeOrderGetter struct {
	options OrderGetOptions
	result  OrderGetResult
}

func (g *fakeOrderGetter) GetOrder(ctx context.Context, options OrderGetOptions) (OrderGetResult, error) {
	g.options = options
	return g.result, nil
}

type fakeOrderBatchGetter struct {
	options OrderBatchGetOptions
	result  OrderBatchGetResult
}

func (g *fakeOrderBatchGetter) BatchGetOrders(ctx context.Context, options OrderBatchGetOptions) (OrderBatchGetResult, error) {
	g.options = options
	return g.result, nil
}

type fakeOrderRefunder struct {
	options OrderRefundOptions
}

func (r *fakeOrderRefunder) RefundOrder(ctx context.Context, options OrderRefundOptions) error {
	r.options = options
	return nil
}
