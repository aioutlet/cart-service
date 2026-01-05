package com.xshopai.cartservice.exception;

public class InsufficientStockException extends CartException {
    
    public InsufficientStockException(String message) {
        super(message);
    }
}
