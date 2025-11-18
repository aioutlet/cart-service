package com.aioutlet.cartservice.client;

import io.dapr.client.DaprClient;
import io.dapr.client.DaprClientBuilder;
import io.dapr.client.domain.HttpExtension;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.enterprise.context.ApplicationScoped;
import org.jboss.logging.Logger;

import jakarta.inject.Inject;
import java.util.Map;

@ApplicationScoped
public class InventoryClient {
    
    @Inject
    Logger logger;
    
    private DaprClient daprClient;
    
    private static final String INVENTORY_SERVICE_APP_ID = "inventory-service";
    
    @PostConstruct
    void init() {
        this.daprClient = new DaprClientBuilder().build();
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
    
    public boolean checkAvailability(String sku, int quantity) {
        try {
            String url = "/api/inventory/check";
            
            logger.infof("Calling inventory service: url=%s, sku=%s, quantity=%d", url, sku, quantity);
            
            // Build request body as inventory service expects
            Map<String, Object> requestBody = Map.of(
                "sku", sku,
                "quantity", quantity
            );
            
            logger.infof("Request body: %s", requestBody);
            
            // Inventory service returns: {"available": true/false, "insufficientItems": [...]}
            @SuppressWarnings("unchecked")
            Map<String, Object> response = daprClient.invokeMethod(
                INVENTORY_SERVICE_APP_ID,
                url,
                requestBody,
                HttpExtension.POST,
                Map.class
            ).block();
            
            logger.infof("Inventory service response: %s", response);
            
            if (response != null && response.containsKey("available")) {
                boolean available = Boolean.TRUE.equals(response.get("available"));
                logger.infof("Inventory availability result: sku=%s, available=%s", sku, available);
                return available;
            }
            
            logger.warnf("Inventory service response missing 'available' field: %s", response);
            return false;
        } catch (Exception e) {
            logger.errorf(e, "Failed to check inventory for SKU %s: %s", sku, e.getMessage());
            throw new RuntimeException("Failed to check inventory: " + sku, e);
        }
    }
}
