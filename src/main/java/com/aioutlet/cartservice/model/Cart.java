package com.xshopai.cartservice.model;

import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

public class Cart {
    
    @JsonProperty("userId")
    private String userId;
    
    @JsonProperty("items")
    private List<CartItem> items = new ArrayList<>();
    
    @JsonProperty("totalPrice")
    private Double totalPrice = 0.0;
    
    @JsonProperty("totalItems")
    private Integer totalItems = 0;
    
    @JsonProperty("createdAt")
    private Long createdAt;
    
    @JsonProperty("updatedAt")
    private Long updatedAt;
    
    @JsonProperty("expiresAt")
    private Long expiresAt;
    
    public Cart() {
        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
    }
    
    public Cart(String userId, Instant expiresAt) {
        this.userId = userId;
        this.items = new ArrayList<>();
        this.totalPrice = 0.0;
        this.totalItems = 0;
        this.createdAt = System.currentTimeMillis();
        this.updatedAt = System.currentTimeMillis();
        this.expiresAt = expiresAt.toEpochMilli();
    }
    
    public void addItem(CartItem item, int maxItems) {
        // Check if item already exists (match by SKU which includes variant info)
        for (CartItem existingItem : items) {
            if (existingItem.getSku().equals(item.getSku())) {
                int newQty = existingItem.getQuantity() + item.getQuantity();
                existingItem.setQuantity(newQty);
                existingItem.setSubtotal(newQty * existingItem.getPrice());
                updateTotals();
                this.updatedAt = System.currentTimeMillis();
                return;
            }
        }
        
        // Check max items limit
        if (items.size() >= maxItems) {
            throw new IllegalArgumentException("Maximum number of items exceeded");
        }
        
        // Add new item
        item.setSubtotal(item.getQuantity() * item.getPrice());
        item.setAddedAt(System.currentTimeMillis());
        items.add(item);
        updateTotals();
        this.updatedAt = System.currentTimeMillis();
    }
    
    public void updateItemQuantity(String sku, int quantity) {
        if (quantity < 0) {
            throw new IllegalArgumentException("Invalid quantity");
        }
        
        items.removeIf(item -> {
            if (item.getSku().equals(sku)) {
                if (quantity == 0) {
                    return true; // Remove item
                }
                item.setQuantity(quantity);
                item.setSubtotal(quantity * item.getPrice());
                return false;
            }
            return false;
        });
        
        updateTotals();
        this.updatedAt = System.currentTimeMillis();
    }
    
    public void removeItem(String sku) {
        items.removeIf(item -> item.getSku().equals(sku));
        updateTotals();
        this.updatedAt = System.currentTimeMillis();
    }
    
    public void clear() {
        items.clear();
        totalPrice = 0.0;
        totalItems = 0;
        this.updatedAt = System.currentTimeMillis();
    }
    
    public boolean isEmpty() {
        return items.isEmpty();
    }
    
    public boolean isExpired() {
        return expiresAt != null && System.currentTimeMillis() > expiresAt;
    }
    
    public void extendExpiry(long hours) {
        this.expiresAt = System.currentTimeMillis() + (hours * 3600 * 1000);
        this.updatedAt = System.currentTimeMillis();
    }
    
    private void updateTotals() {
        totalPrice = items.stream()
                .mapToDouble(CartItem::getSubtotal)
                .sum();
        totalItems = items.stream()
                .mapToInt(CartItem::getQuantity)
                .sum();
    }
    
    // Getters and Setters
    public String getUserId() {
        return userId;
    }
    
    public void setUserId(String userId) {
        this.userId = userId;
    }
    
    public List<CartItem> getItems() {
        return items;
    }
    
    public void setItems(List<CartItem> items) {
        this.items = items;
        updateTotals();
    }
    
    public Double getTotalPrice() {
        return totalPrice;
    }
    
    public void setTotalPrice(Double totalPrice) {
        this.totalPrice = totalPrice;
    }
    
    public Integer getTotalItems() {
        return totalItems;
    }
    
    public void setTotalItems(Integer totalItems) {
        this.totalItems = totalItems;
    }
    
    public Long getCreatedAt() {
        return createdAt;
    }
    
    public void setCreatedAt(Long createdAt) {
        this.createdAt = createdAt;
    }
    
    public Long getUpdatedAt() {
        return updatedAt;
    }
    
    public void setUpdatedAt(Long updatedAt) {
        this.updatedAt = updatedAt;
    }
    
    public Long getExpiresAt() {
        return expiresAt;
    }
    
    public void setExpiresAt(Long expiresAt) {
        this.expiresAt = expiresAt;
    }
}
