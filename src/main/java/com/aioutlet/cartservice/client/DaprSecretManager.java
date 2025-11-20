package com.aioutlet.cartservice.client;

import io.dapr.client.DaprClient;
import io.dapr.client.DaprClientBuilder;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.enterprise.context.ApplicationScoped;
import jakarta.inject.Inject;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.jboss.logging.Logger;

import java.util.Map;

/**
 * Dapr Secret Manager
 * Handles retrieving secrets from Dapr secret store building block
 * 
 * This class uses the Dapr SDK to securely retrieve secrets from the configured
 * secret store (local file, Azure Key Vault, AWS Secrets Manager, etc.)
 */
@ApplicationScoped
public class DaprSecretManager {

    @Inject
    Logger logger;

    private static final String SECRET_STORE_NAME = "secret-store";

    private DaprClient daprClient;

    @PostConstruct
    void init() {
        this.daprClient = new DaprClientBuilder().build();
        logger.infof("Dapr Secret Manager initialized with store: %s", SECRET_STORE_NAME);
    }

    @PreDestroy
    void cleanup() {
        if (daprClient != null) {
            try {
                daprClient.close();
                logger.info("Dapr client closed successfully");
            } catch (Exception e) {
                logger.error("Error closing Dapr client", e);
            }
        }
    }

    /**
     * Get a specific secret by key
     * Supports nested keys with colon separator (e.g., "database:password")
     * 
     * @param key The secret key to retrieve
     * @return The secret value, or null if not found
     */
    public String getSecret(String key) {
        try {
            logger.debugf("Retrieving secret: %s", key);
            
            Map<String, String> secret = daprClient.getSecret(SECRET_STORE_NAME, key).block();
            
            if (secret == null || secret.isEmpty()) {
                logger.warnf("Secret not found: %s", key);
                return null;
            }
            
            // Return the first value (Dapr handles nested separator automatically)
            String value = secret.values().stream().findFirst().orElse(null);
            logger.debugf("Successfully retrieved secret: %s", key);
            return value;
        } catch (Exception e) {
            logger.errorf("Failed to retrieve secret: %s", key, e);
            throw new RuntimeException("Failed to retrieve secret: " + key, e);
        }
    }

    /**
     * Get all secrets for a specific key (returns all metadata)
     * 
     * @param key The secret key to retrieve
     * @return Map of secret metadata
     */
    public Map<String, String> getSecrets(String key) {
        try {
            logger.debugf("Retrieving secrets for key: %s", key);
            return daprClient.getSecret(SECRET_STORE_NAME, key).block();
        } catch (Exception e) {
            logger.errorf("Failed to retrieve secrets for key: %s", key, e);
            throw new RuntimeException("Failed to retrieve secrets for key: " + key, e);
        }
    }

    /**
     * Get Redis password for state store
     * 
     * @return Redis password
     */
    public String getRedisPassword() {
        return getSecret("redis:password");
    }

    /**
     * Get JWT secret
     * Uses nested structure: jwt:secret
     * 
     * @return JWT secret
     */
    public String getJwtSecret() {
        return getSecret("jwt:secret");
    }
}
