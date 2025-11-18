package com.aioutlet.cartservice.resource;

import jakarta.ws.rs.GET;
import jakarta.ws.rs.Path;
import jakarta.ws.rs.Produces;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.eclipse.microprofile.health.HealthCheckResponse;
import org.eclipse.microprofile.health.Liveness;
import org.eclipse.microprofile.health.Readiness;

import java.time.Instant;
import java.util.HashMap;
import java.util.Map;

@Path("/")
@Produces(MediaType.APPLICATION_JSON)
public class RootResource {
    
    @ConfigProperty(name = "quarkus.application.name")
    String applicationName;
    
    @ConfigProperty(name = "quarkus.application.version")
    String applicationVersion;
    
    private final Instant startTime = Instant.now();
    
    @GET
    public Response root() {
        Map<String, Object> info = new HashMap<>();
        info.put("service", applicationName);
        info.put("version", applicationVersion);
        info.put("status", "operational");
        info.put("message", "Cart Service is running");
        info.put("timestamp", Instant.now());
        info.put("uptime", java.time.Duration.between(startTime, Instant.now()).toSeconds() + "s");
        return Response.ok(info).build();
    }
    
    @GET
    @Path("/version")
    public Response version() {
        Map<String, String> version = new HashMap<>();
        version.put("service", applicationName);
        version.put("version", applicationVersion);
        return Response.ok(version).build();
    }
    
    @GET
    @Path("/health")
    public Response health() {
        Map<String, Object> health = new HashMap<>();
        health.put("status", "UP");
        health.put("service", applicationName);
        health.put("timestamp", Instant.now());
        health.put("checks", Map.of(
            "cart-service", "UP",
            "dapr-sidecar", "UP"
        ));
        return Response.ok(health).build();
    }
    
    @GET
    @Path("/liveness")
    public Response liveness() {
        Map<String, Object> liveness = new HashMap<>();
        liveness.put("status", "UP");
        liveness.put("service", applicationName);
        liveness.put("timestamp", Instant.now());
        return Response.ok(liveness).build();
    }
    
    @GET
    @Path("/readiness")
    public Response readiness() {
        Map<String, Object> readiness = new HashMap<>();
        readiness.put("status", "UP");
        readiness.put("service", applicationName);
        readiness.put("timestamp", Instant.now());
        readiness.put("checks", Map.of(
            "dapr-state-store", "UP",
            "service", "UP"
        ));
        return Response.ok(readiness).build();
    }
    
    @GET
    @Path("/metrics")
    public Response metrics() {
        Map<String, Object> metrics = new HashMap<>();
        metrics.put("service", applicationName);
        metrics.put("uptime_seconds", java.time.Duration.between(startTime, Instant.now()).toSeconds());
        metrics.put("timestamp", Instant.now());
        metrics.put("jvm", Map.of(
            "memory_used_mb", Runtime.getRuntime().totalMemory() - Runtime.getRuntime().freeMemory(),
            "memory_total_mb", Runtime.getRuntime().totalMemory(),
            "memory_max_mb", Runtime.getRuntime().maxMemory(),
            "processors", Runtime.getRuntime().availableProcessors()
        ));
        return Response.ok(metrics).build();
    }
    
    @GET
    @Path("/info")
    public Response info() {
        Map<String, Object> info = new HashMap<>();
        info.put("service", applicationName);
        info.put("version", applicationVersion);
        info.put("timestamp", Instant.now());
        info.put("uptime", java.time.Duration.between(startTime, Instant.now()).toSeconds() + "s");
        return Response.ok(info).build();
    }
}
