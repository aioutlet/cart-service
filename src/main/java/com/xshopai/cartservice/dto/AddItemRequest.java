package com.xshopai.cartservice.dto;

import jakarta.validation.constraints.Min;
import jakarta.validation.constraints.NotNull;

public class AddItemRequest {
    
    @NotNull(message = "Product ID is required")
    private String productId;
    
    @NotNull(message = "Quantity is required")
    @Min(value = 1, message = "Quantity must be at least 1")
    private Integer quantity;
    
    private String imageUrl;
    private String selectedColor;
    private String selectedSize;
    
    public AddItemRequest() {
    }
    
    public AddItemRequest(String productId, Integer quantity) {
        this.productId = productId;
        this.quantity = quantity;
    }
    
    // Getters and Setters
    public String getProductId() {
        return productId;
    }
    
    public void setProductId(String productId) {
        this.productId = productId;
    }
    
    public Integer getQuantity() {
        return quantity;
    }
    
    public void setQuantity(Integer quantity) {
        this.quantity = quantity;
    }
    
    public String getImageUrl() {
        return imageUrl;
    }
    
    public void setImageUrl(String imageUrl) {
        this.imageUrl = imageUrl;
    }
    
    public String getSelectedColor() {
        return selectedColor;
    }
    
    public void setSelectedColor(String selectedColor) {
        this.selectedColor = selectedColor;
    }
    
    public String getSelectedSize() {
        return selectedSize;
    }
    
    public void setSelectedSize(String selectedSize) {
        this.selectedSize = selectedSize;
    }
}
