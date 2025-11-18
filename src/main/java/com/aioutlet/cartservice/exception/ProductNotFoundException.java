package com.aioutlet.cartservice.exception;

public class ProductNotFoundException extends CartException {
    
    public ProductNotFoundException(String message) {
        super(message);
    }
}
