package cart

import (
	"ecom/types"
	"fmt"
)

func getCartItemsIDs(items []types.CartItem) ([]int, error) {
	productsIds := make([]int, len(items))
	for i, item := range items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("invalid quantity for product %d", item.ProductId)
		}
		productsIds[i] = item.ProductId
	}
	return productsIds, nil
}

func (h *Handler) createOrder(ps []types.Product, items []types.CartItem, userID int) (int, float64, error) {
	productMap := make(map[int]types.Product)
	for _, product := range ps {
		productMap[product.ID] = product
	}
	if err := checkIfCartIsInStock(items, productMap); err != nil {
		return 0, 0, nil
	}

	totalPrice := calculateTotalPrice(items, productMap)

	for _, item := range items {
		product := productMap[item.ProductId]
		product.Quantity -= item.Quantity

		h.productStore.UpdateProduct(product)
	}

	orderID, err := h.store.CreateOrder(types.Order{
		UserID:  userID,
		Total:   totalPrice,
		Status:  "pending",
		Address: "some Address",
	})
	if err != nil {
		return 0, 0, nil
	}

	for _, item := range items {
		h.store.CreateOrderItem(types.OrderItem{
			OrderID:   orderID,
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			Price:     productMap[item.ProductId].Price,
		})
	}

	return orderID, totalPrice, nil
}

func checkIfCartIsInStock(cartItems []types.CartItem, product map[int]types.Product) error {
	if len(cartItems) == 0 {
		return fmt.Errorf("cart is empty")
	}

	for _, item := range cartItems {
		product, ok := product[item.ProductId]
		if !ok {
			return fmt.Errorf("Product %d is not available in the store, please refresh your cart", item.ProductId)
		}
		if product.Quantity < item.Quantity {
			return fmt.Errorf("Product %d is not available in the quantity requested", product.Name)
		}
	}

	return nil
}

func calculateTotalPrice(cartItems []types.CartItem, product map[int]types.Product) float64 {
	var total float64

	for _, item := range cartItems {
		product := product[item.ProductId]
		total += product.Price * float64(item.Quantity)
	}

	return total
}
