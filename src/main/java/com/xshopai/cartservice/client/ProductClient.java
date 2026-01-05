package com.xshopai.cartservice.client;

import com.xshopai.cartservice.model.ProductInfo;
import io.dapr.client.DaprClient;
import io.dapr.client.DaprClientBuilder;
import io.dapr.client.domain.HttpExtension;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.enterprise.context.ApplicationScoped;
import org.jboss.logging.Logger;

import jakarta.inject.Inject;

@ApplicationScoped
public class ProductClient {
    
    @Inject
    Logger logger;
    
    private DaprClient daprClient;
    
    private static final String PRODUCT_SERVICE_APP_ID = "product-service";
    
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
    
    public ProductInfo getProduct(String productId) {
        try {
            ProductInfo product = daprClient.invokeMethod(
                PRODUCT_SERVICE_APP_ID,
                "/api/products/" + productId,
                null,
                HttpExtension.GET,
                ProductInfo.class
            ).block();
            
            return product;
        } catch (Exception e) {
            logger.errorf("Failed to get product %s: %s", productId, e.getMessage());
            throw new RuntimeException("Failed to get product: " + productId, e);
        }
    }
}
