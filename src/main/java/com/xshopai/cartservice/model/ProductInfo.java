package com.xshopai.cartservice.model;

import com.fasterxml.jackson.annotation.JsonProperty;

public class ProductInfo {
    
    @JsonProperty("id")
    private String id;
    
    @JsonProperty("name")
    private String name;
    
    @JsonProperty("sku")
    private String sku;
    
    @JsonProperty("price")
    private Double price;
    
    @JsonProperty("imageUrl")
    private String imageUrl;
    
    @JsonProperty("category")
    private String category;
    
    @JsonProperty("is_active")
    private Boolean isActive;
    
    @JsonProperty("stockQty")
    private Integer stockQty;
    
    public ProductInfo() {
    }
    
    // Getters and Setters
    public String getId() {
        return id;
    }
    
    public void setId(String id) {
        this.id = id;
    }
    
    public String getName() {
        return name;
    }
    
    public void setName(String name) {
        this.name = name;
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
    
    public Boolean getIsActive() {
        return isActive;
    }
    
    public void setIsActive(Boolean isActive) {
        this.isActive = isActive;
    }
    
    public Integer getStockQty() {
        return stockQty;
    }
    
    public void setStockQty(Integer stockQty) {
        this.stockQty = stockQty;
    }
}
