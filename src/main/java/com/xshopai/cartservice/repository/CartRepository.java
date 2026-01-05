package com.xshopai.cartservice.repository;

import com.xshopai.cartservice.model.Cart;
import io.dapr.client.DaprClient;
import io.dapr.client.DaprClientBuilder;
import io.dapr.client.domain.State;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.jboss.logging.Logger;

import java.time.Duration;
import java.util.Optional;

@ApplicationScoped
public class CartRepository {
    
    @Inject
    Logger logger;
    
    @ConfigProperty(name = "dapr.state-store", defaultValue = "statestore")
    String stateStoreName;
    
    private DaprClient daprClient;
    
    private static final String CART_PREFIX = "cart:";
    private static final String LOCK_PREFIX = "lock:cart:";
    
    @PostConstruct
    void init() {
        // Note: Dapr SDK's DefaultObjectSerializer creates its own ObjectMapper
        // We'll use the default serializer which should work with JSR310 module on classpath
        this.daprClient = new DaprClientBuilder().build();
        logger.info("Dapr client initialized for state store: " + stateStoreName);
    }
    
    @PreDestroy
    void cleanup() {
        try {
            if (daprClient != null) {
                daprClient.close();
            }
        } catch (Exception e) {
            logger.error("Error closing Dapr client", e);
        }
    }
    
    public Optional<Cart> findByUserId(String userId) {
        try {
            State<Cart> state = daprClient.getState(
                stateStoreName, 
                CART_PREFIX + userId, 
                Cart.class
            ).block();
            
            if (state != null && state.getValue() != null) {
                return Optional.of(state.getValue());
            }
            return Optional.empty();
        } catch (Exception e) {
            logger.error("Error finding cart for user: " + userId, e);
            return Optional.empty();
        }
    }
    
    public void save(Cart cart, Duration ttl) {
        try {
            // Save cart object directly - Dapr SDK handles JSON serialization
            daprClient.saveState(
                stateStoreName,
                CART_PREFIX + cart.getUserId(),
                cart
            ).block();
            
            logger.debugf("Saved cart for user: %s", cart.getUserId());
        } catch (Exception e) {
            logger.error("Error saving cart for user: " + cart.getUserId(), e);
            throw new RuntimeException("Failed to save cart", e);
        }
    }
    
    public void delete(String userId) {
        try {
            daprClient.deleteState(stateStoreName, CART_PREFIX + userId).block();
            logger.debugf("Deleted cart for user: %s", userId);
        } catch (Exception e) {
            logger.error("Error deleting cart for user: " + userId, e);
        }
    }
    
    public boolean acquireLock(String userId, Duration lockDuration) {
        try {
            // Check if lock already exists
            State<String> existingLock = daprClient.getState(
                stateStoreName,
                LOCK_PREFIX + userId,
                String.class
            ).block();
            
            if (existingLock != null && existingLock.getValue() != null && !existingLock.getValue().isEmpty()) {
                logger.debugf("Failed to acquire lock for user: %s - already locked", userId);
                return false;
            }
            
            // Acquire lock (note: TTL metadata not supported in Dapr SDK 1.11.0)
            daprClient.saveState(
                stateStoreName,
                LOCK_PREFIX + userId,
                "locked"
            ).block();
            
            logger.debugf("Acquired lock for user: %s", userId);
            return true;
        } catch (Exception e) {
            logger.error("Error acquiring lock for user: " + userId, e);
            return false;
        }
    }
    
    public void releaseLock(String userId) {
        try {
            daprClient.deleteState(stateStoreName, LOCK_PREFIX + userId).block();
            logger.debugf("Released lock for user: %s", userId);
        } catch (Exception e) {
            logger.error("Error releasing lock for user: " + userId, e);
        }
    }
}
