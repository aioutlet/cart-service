package com.aioutlet.cartservice.dto;

import jakarta.validation.constraints.NotBlank;

public class TransferCartRequest {
    
    @NotBlank(message = "Guest ID is required")
    private String guestId;
    
    public TransferCartRequest() {
    }
    
    public TransferCartRequest(String guestId) {
        this.guestId = guestId;
    }
    
    public String getGuestId() {
        return guestId;
    }
    
    public void setGuestId(String guestId) {
        this.guestId = guestId;
    }
}
