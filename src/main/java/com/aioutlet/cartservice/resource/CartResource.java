package com.aioutlet.cartservice.resource;

import com.aioutlet.cartservice.dto.AddItemRequest;
import com.aioutlet.cartservice.dto.CartResponse;
import com.aioutlet.cartservice.dto.TransferCartRequest;
import com.aioutlet.cartservice.dto.UpdateItemRequest;
import com.aioutlet.cartservice.model.Cart;
import com.aioutlet.cartservice.service.CartService;
import jakarta.inject.Inject;
import jakarta.validation.Valid;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import org.eclipse.microprofile.openapi.annotations.Operation;
import org.eclipse.microprofile.openapi.annotations.tags.Tag;
import org.jboss.logging.Logger;

@Path("/api/v1")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
@Tag(name = "Cart", description = "Shopping cart operations")
public class CartResource {
    
    @Inject
    Logger logger;
    
    @Inject
    CartService cartService;
    
    // Authenticated Cart Endpoints
    
    @GET
    @Path("/cart")
    @Operation(summary = "Get user cart", description = "Retrieve the authenticated user's shopping cart")
    public Response getCart(@HeaderParam("X-User-ID") String userId) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            Cart cart = cartService.getCart(userId);
            return Response.ok(CartResponse.success("Cart retrieved successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error getting cart", e);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(CartResponse.error("Failed to retrieve cart"))
                .build();
        }
    }
    
    @POST
    @Path("/cart/items")
    @Operation(summary = "Add item to cart", description = "Add an item to the authenticated user's cart")
    public Response addItem(@HeaderParam("X-User-ID") String userId, @Valid AddItemRequest request) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            Cart cart = cartService.addItem(userId, request, false);
            return Response.ok(CartResponse.success("Item added to cart successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error adding item to cart", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @PUT
    @Path("/cart/items/{sku}")
    @Operation(summary = "Update item quantity", description = "Update the quantity of an item in the cart by SKU")
    public Response updateItem(@HeaderParam("X-User-ID") String userId,
                              @PathParam("sku") String sku,
                              @Valid UpdateItemRequest request) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            Cart cart = cartService.updateItemQuantity(userId, sku, request.getQuantity());
            return Response.ok(CartResponse.success("Item updated successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error updating cart item", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @DELETE
    @Path("/cart/items/{sku}")
    @Operation(summary = "Remove item from cart", description = "Remove an item from the cart by SKU")
    public Response removeItem(@HeaderParam("X-User-ID") String userId,
                              @PathParam("sku") String sku) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            Cart cart = cartService.removeItem(userId, sku);
            return Response.ok(CartResponse.success("Item removed successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error removing cart item", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @DELETE
    @Path("/cart")
    @Operation(summary = "Clear cart", description = "Remove all items from the cart")
    public Response clearCart(@HeaderParam("X-User-ID") String userId) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            cartService.clearCart(userId);
            return Response.ok(CartResponse.success("Cart cleared successfully", null)).build();
        } catch (Exception e) {
            logger.error("Error clearing cart", e);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(CartResponse.error("Failed to clear cart"))
                .build();
        }
    }
    
    @POST
    @Path("/cart/transfer")
    @Operation(summary = "Transfer guest cart", description = "Transfer guest cart to authenticated user")
    public Response transferCart(@HeaderParam("X-User-ID") String userId,
                                 @Valid TransferCartRequest request) {
        if (userId == null || userId.isEmpty()) {
            return Response.status(Response.Status.UNAUTHORIZED)
                .entity(CartResponse.error("User not authenticated"))
                .build();
        }
        
        try {
            Cart cart = cartService.transferCart(request.getGuestId(), userId);
            return Response.ok(CartResponse.success("Cart transferred successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error transferring cart", e);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(CartResponse.error("Failed to transfer cart"))
                .build();
        }
    }
    
    // Guest Cart Endpoints
    
    @GET
    @Path("/guest/cart/{guestId}")
    @Operation(summary = "Get guest cart", description = "Retrieve a guest user's shopping cart")
    public Response getGuestCart(@PathParam("guestId") String guestId) {
        try {
            Cart cart = cartService.getCart(guestId);
            return Response.ok(CartResponse.success("Cart retrieved successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error getting guest cart", e);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(CartResponse.error("Failed to retrieve cart"))
                .build();
        }
    }
    
    @POST
    @Path("/guest/cart/{guestId}/items")
    @Operation(summary = "Add item to guest cart", description = "Add an item to a guest user's cart")
    public Response addGuestItem(@PathParam("guestId") String guestId, @Valid AddItemRequest request) {
        try {
            Cart cart = cartService.addItem(guestId, request, true);
            return Response.ok(CartResponse.success("Item added to cart successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error adding item to guest cart", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @PUT
    @Path("/guest/cart/{guestId}/items/{sku}")
    @Operation(summary = "Update guest cart item", description = "Update item quantity in guest cart by SKU")
    public Response updateGuestItem(@PathParam("guestId") String guestId,
                                   @PathParam("sku") String sku,
                                   @Valid UpdateItemRequest request) {
        try {
            Cart cart = cartService.updateItemQuantity(guestId, sku, request.getQuantity());
            return Response.ok(CartResponse.success("Item updated successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error updating guest cart item", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @DELETE
    @Path("/guest/cart/{guestId}/items/{sku}")
    @Operation(summary = "Remove item from guest cart", description = "Remove an item from guest cart by SKU")
    public Response removeGuestItem(@PathParam("guestId") String guestId,
                                   @PathParam("sku") String sku) {
        try {
            Cart cart = cartService.removeItem(guestId, sku);
            return Response.ok(CartResponse.success("Item removed successfully", cart)).build();
        } catch (Exception e) {
            logger.error("Error removing guest cart item", e);
            return Response.status(Response.Status.BAD_REQUEST)
                .entity(CartResponse.error(e.getMessage()))
                .build();
        }
    }
    
    @DELETE
    @Path("/guest/cart/{guestId}")
    @Operation(summary = "Clear guest cart", description = "Remove all items from guest cart")
    public Response clearGuestCart(@PathParam("guestId") String guestId) {
        try {
            cartService.clearCart(guestId);
            return Response.ok(CartResponse.success("Cart cleared successfully", null)).build();
        } catch (Exception e) {
            logger.error("Error clearing guest cart", e);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(CartResponse.error("Failed to clear cart"))
                .build();
        }
    }
}
