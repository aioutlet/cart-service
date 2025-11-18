package com.aioutlet.cartservice.exception;

public class CartException extends RuntimeException {
    
    public CartException(String message) {
        super(message);
    }
    
    public CartException(String message, Throwable cause) {
        super(message, cause);
    }
}
