import React from "react";
import { useNavigate } from "react-router-dom";

function RegionCard({ name, macrozoneCount }) {
    const navigate = useNavigate();

    const handleClick = () => {
        navigate(`/region/${encodeURIComponent(name)}`);
    };

    return (
        <div className="region-card" onClick={handleClick} tabIndex={0} role="button">
            <div className="region-card-header">
                <span className="region-card-icon">
                    <svg width="22" height="22" fill="#fff" viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="10" fill="#1976d2" />
                        <text x="12" y="16" textAnchor="middle" fontSize="12" fill="#fff">R</text>
                    </svg>
                </span>
                <h2>{name}</h2>
            </div>
            <p>Macrozone: <b>{macrozoneCount}</b></p>
        </div>
    );
}

export default RegionCard;