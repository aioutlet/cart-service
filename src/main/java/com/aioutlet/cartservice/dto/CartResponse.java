package com.aioutlet.cartservice.dto;

public class CartResponse {
    private boolean success;
    private String message;
    private Object data;
    
    public CartResponse() {
    }
    
    public CartResponse(boolean success, String message, Object data) {
        this.success = success;
        this.message = message;
        this.data = data;
    }
    
    public static CartResponse success(String message, Object data) {
        return new CartResponse(true, message, data);
    }
    
    public static CartResponse error(String message) {
        return new CartResponse(false, message, null);
    }
    
    // Getters and Setters
    public boolean isSuccess() {
        return success;
    }
    
    public void setSuccess(boolean success) {
        this.success = success;
    }
    
    public String getMessage() {
        return message;
    }
    
    public void setMessage(String message) {
        this.message = message;
    }
    
    public Object getData() {
        return data;
    }
    
    public void setData(Object data) {
        this.data = data;
    }
}
