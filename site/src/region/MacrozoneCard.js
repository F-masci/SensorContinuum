import React from "react";
import { useNavigate, useParams } from "react-router-dom";

function MacrozoneCard({ name, zoneCount, lat, lon, creationTime }) {
    const navigate = useNavigate();
    const { name: regionName } = useParams();

    const handleClick = () => {
        navigate(`/macrozone/${encodeURIComponent(regionName)}/${encodeURIComponent(name)}`);
    };

    return (
        <div className="macrozone-card" onClick={handleClick} tabIndex={0} role="button">
            <div className="macrozone-card-header">
                <span className="macrozone-card-icon">
                    <svg width="22" height="22" fill="#fff" viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="10" fill="#1976d2" />
                        <text x="12" y="16" textAnchor="middle" fontSize="12" fill="#fff">MZ</text>
                    </svg>
                </span>
                <h2>{name}</h2>
            </div>
            <p>Zone: <b>{zoneCount}</b></p>
            <p>Lat: {lat}, Lon: {lon}</p>
            <p>Creato il: {new Date(creationTime).toLocaleString()}</p>
        </div>
    );
}

export default MacrozoneCard;