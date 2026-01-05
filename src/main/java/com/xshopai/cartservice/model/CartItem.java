package com.xshopai.cartservice.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public class CartItem {
    
    @JsonProperty("productId")
    private String productId;
    
    @JsonProperty("productName")
    private String productName;
    
    @JsonProperty("sku")
    private String sku;
    
    @JsonProperty("price")
    private Double price;
    
    @JsonProperty("quantity")
    private Integer quantity;
    
    @JsonProperty("imageUrl")
    private String imageUrl;
    
    @JsonProperty("category")
    private String category;
    
    @JsonProperty("subtotal")
    private Double subtotal;
    
    @JsonProperty("selectedColor")
    private String selectedColor;
    
    @JsonProperty("selectedSize")
    private String selectedSize;
    
    @JsonProperty("addedAt")
    private Long addedAt;
    
    public CartItem() {
    }
    
    public CartItem(String productId, String productName, String sku, Double price, 
                    Integer quantity, String imageUrl, String category) {
        this.productId = productId;
        this.productName = productName;
        this.sku = sku;
        this.price = price;
        this.quantity = quantity;
        this.imageUrl = imageUrl;
        this.category = category;
        this.subtotal = price * quantity;
        this.addedAt = System.currentTimeMillis();
    }
    
    // Getters and Setters
    public String getProductId() {
        return productId;
    }
    
    public void setProductId(String productId) {
        this.productId = productId;
    }
    
    public String getProductName() {
        return productName;
    }
    
    public void setProductName(String productName) {
        this.productName = productName;
    }
    
    public String getSku() {
        return sku;
    }
    
    public void setSku(String sku) {
        this.sku = sku;
    }
    
    public Double getPrice() {
        return price;
    }
    
    public void setPrice(Double price) {
        this.price = price;
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
    
    public String getCategory() {
        return category;
    }
    
    public void setCategory(String category) {
        this.category = category;
    }
    
    public Double getSubtotal() {
        return subtotal;
    }
    
    public void setSubtotal(Double subtotal) {
        this.subtotal = subtotal;
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
    
    public Long getAddedAt() {
        return addedAt;
    }
    
    public void setAddedAt(Long addedAt) {
        this.addedAt = addedAt;
    }
}
