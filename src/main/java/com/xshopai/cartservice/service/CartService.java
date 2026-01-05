package com.xshopai.cartservice.service;

import com.xshopai.cartservice.client.InventoryClient;
import com.xshopai.cartservice.client.ProductClient;
import com.xshopai.cartservice.dto.AddItemRequest;
import com.xshopai.cartservice.exception.CartException;
import com.xshopai.cartservice.exception.InsufficientStockException;
import com.xshopai.cartservice.exception.ProductNotFoundException;
import com.xshopai.cartservice.model.Cart;
import com.xshopai.cartservice.model.CartItem;
import com.xshopai.cartservice.model.ProductInfo;
import com.xshopai.cartservice.repository.CartRepository;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.eclipse.microprofile.rest.client.inject.RestClient;
import org.jboss.logging.Logger;

import java.time.Duration;
import java.time.Instant;

@ApplicationScoped
public class CartService {
    
    @Inject
    Logger logger;
    
    @Inject
    CartRepository cartRepository;
    
    @Inject
    ProductClient productClient;
    
    @Inject
    InventoryClient inventoryClient;
    
    @ConfigProperty(name = "cart.default-ttl", defaultValue = "720h")
    String defaultTtlConfig;
    
    @ConfigProperty(name = "cart.guest-ttl", defaultValue = "72h")
    String guestTtlConfig;
    
    @ConfigProperty(name = "cart.max-items", defaultValue = "100")
    int maxItems;
    
    public Cart getCart(String userId) {
        return cartRepository.findByUserId(userId)
            .orElseGet(() -> createNewCart(userId, false));
    }
    
    public Cart addItem(String userId, AddItemRequest request, boolean isGuest) {
        // Acquire lock
        if (!cartRepository.acquireLock(userId, Duration.ofSeconds(30))) {
            throw new CartException("Cart is currently being modified, please try again");
        }
        
        try {
            // Get product info
            ProductInfo productInfo;
            try {
                productInfo = productClient.getProduct(request.getProductId());
            } catch (Exception e) {
                logger.errorf("Failed to get product %s: %s", request.getProductId(), e.getMessage());
                throw new ProductNotFoundException("Product not found: " + request.getProductId());
            }
            
            if (productInfo == null || productInfo.getId() == null) {
                throw new ProductNotFoundException("Product not found: " + request.getProductId());
            }
            
            // Check if product is active
            if (productInfo.getIsActive() != null && !productInfo.getIsActive()) {
                throw new CartException("Product is not available");
            }
            
            // Generate variant SKU
            String variantSku = generateVariantSku(productInfo.getSku(), 
                request.getSelectedColor(), request.getSelectedSize());
            
            logger.infof("Generated variant SKU: %s (base: %s, color: %s, size: %s)", 
                variantSku, productInfo.getSku(), request.getSelectedColor(), request.getSelectedSize());
            
            // Check inventory (non-blocking, log warning if fails)
            try {
                logger.infof("Checking inventory availability: sku=%s, quantity=%d", 
                    variantSku, request.getQuantity());
                
                boolean available = inventoryClient.checkAvailability(variantSku, request.getQuantity());
                
                logger.infof("Inventory check result: sku=%s, available=%s", variantSku, available);
                
                if (!available) {
                    logger.warnf("Insufficient stock for SKU %s, quantity %d", variantSku, request.getQuantity());
                    throw new InsufficientStockException("Insufficient stock for product");
                }
            } catch (InsufficientStockException e) {
                logger.errorf("InsufficientStockException: %s", e.getMessage());
                throw e;
            } catch (Exception e) {
                logger.warnf("Failed to check inventory for SKU %s, allowing operation: %s", 
                    variantSku, e.getMessage());
            }
            
            // Get or create cart
            Cart cart = getCart(userId);
            
            // Use imageUrl from request if product service doesn't provide one
            String imageUrl = productInfo.getImageUrl();
            if (imageUrl == null || imageUrl.isEmpty()) {
                imageUrl = request.getImageUrl();
            }
            if (imageUrl == null || imageUrl.isEmpty()) {
                imageUrl = "";
            }
            
            // Create cart item
            CartItem cartItem = new CartItem();
            cartItem.setProductId(productInfo.getId());
            cartItem.setProductName(productInfo.getName());
            cartItem.setSku(variantSku);
            cartItem.setPrice(productInfo.getPrice());
            cartItem.setQuantity(request.getQuantity());
            cartItem.setImageUrl(imageUrl);
            cartItem.setCategory(productInfo.getCategory() != null ? productInfo.getCategory() : "");
            cartItem.setSelectedColor(request.getSelectedColor());
            cartItem.setSelectedSize(request.getSelectedSize());
            cartItem.setAddedAt(System.currentTimeMillis());
            
            // Add item to cart
            cart.addItem(cartItem, maxItems);
            
            // Save cart with appropriate TTL
            Duration ttl = isGuest ? parseDuration(guestTtlConfig) : parseDuration(defaultTtlConfig);
            cartRepository.save(cart, ttl);
            
            logger.infof("Item added to cart: userId=%s, productId=%s, quantity=%d", 
                userId, request.getProductId(), request.getQuantity());
            
            return cart;
        } finally {
            cartRepository.releaseLock(userId);
        }
    }
    
    public Cart updateItemQuantity(String userId, String sku, int quantity) {
        if (!cartRepository.acquireLock(userId, Duration.ofSeconds(30))) {
            throw new CartException("Cart is currently being modified, please try again");
        }
        
        try {
            Cart cart = getCart(userId);
            cart.updateItemQuantity(sku, quantity);
            cartRepository.save(cart, parseDuration(defaultTtlConfig));
            logger.infof("Item quantity updated: userId=%s, sku=%s, quantity=%d", 
                userId, sku, quantity);
            return cart;
        } finally {
            cartRepository.releaseLock(userId);
        }
    }
    
    public Cart removeItem(String userId, String sku) {
        if (!cartRepository.acquireLock(userId, Duration.ofSeconds(30))) {
            throw new CartException("Cart is currently being modified, please try again");
        }
        
        try {
            Cart cart = getCart(userId);
            cart.removeItem(sku);
            cartRepository.save(cart, parseDuration(defaultTtlConfig));
            logger.infof("Item removed from cart: userId=%s, sku=%s", userId, sku);
            return cart;
        } finally {
            cartRepository.releaseLock(userId);
        }
    }
    
    public void clearCart(String userId) {
        cartRepository.delete(userId);
        logger.infof("Cart cleared: userId=%s", userId);
    }
    
    public Cart transferCart(String guestId, String userId) {
        Cart guestCart = cartRepository.findByUserId(guestId).orElse(null);
        if (guestCart == null || guestCart.isEmpty()) {
            logger.infof("No guest cart to transfer: guestId=%s", guestId);
            return getCart(userId);
        }
        
        Cart userCart = getCart(userId);
        
        // Transfer items
        int transferredCount = 0;
        for (CartItem item : guestCart.getItems()) {
            try {
                userCart.addItem(item, maxItems);
                transferredCount++;
            } catch (Exception e) {
                logger.warnf("Failed to transfer item %s: %s", item.getProductId(), e.getMessage());
            }
        }
        
        cartRepository.save(userCart, parseDuration(defaultTtlConfig));
        cartRepository.delete(guestId);
        
        logger.infof("Cart transferred: guestId=%s, userId=%s, items=%d", 
            guestId, userId, transferredCount);
        
        return userCart;
    }
    
    private Cart createNewCart(String userId, boolean isGuest) {
        Duration ttl = isGuest ? parseDuration(guestTtlConfig) : parseDuration(defaultTtlConfig);
        Instant expiresAt = Instant.now().plus(ttl);
        Cart cart = new Cart(userId, expiresAt);
        logger.debugf("Created new cart: userId=%s, isGuest=%b", userId, isGuest);
        return cart;
    }
    
    private String generateVariantSku(String baseSku, String color, String size) {
        if (baseSku == null) {
            baseSku = "UNKNOWN";
        }
        
        StringBuilder sku = new StringBuilder(baseSku);
        if (color != null && !color.isEmpty()) {
            sku.append("-").append(color.toUpperCase());
        }
        if (size != null && !size.isEmpty()) {
            sku.append("-").append(size.toUpperCase());
        }
        return sku.toString();
    }
    
    private Duration parseDuration(String durationStr) {
        try {
            if (durationStr.endsWith("h")) {
                long hours = Long.parseLong(durationStr.substring(0, durationStr.length() - 1));
                return Duration.ofHours(hours);
            } else if (durationStr.endsWith("m")) {
                long minutes = Long.parseLong(durationStr.substring(0, durationStr.length() - 1));
                return Duration.ofMinutes(minutes);
            } else if (durationStr.endsWith("s")) {
                long seconds = Long.parseLong(durationStr.substring(0, durationStr.length() - 1));
                return Duration.ofSeconds(seconds);
            }
            // Default to hours
            return Duration.ofHours(Long.parseLong(durationStr));
        } catch (Exception e) {
            logger.warnf("Failed to parse duration '%s', using default 24h", durationStr);
            return Duration.ofHours(24);
        }
    }
}
