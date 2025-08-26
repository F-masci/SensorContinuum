import React from "react";

import './SensorCard.css'

function isSensorActive(lastSeen) {
    const now = new Date();
    const lastSeenDate = new Date(lastSeen);
    return (now - lastSeenDate) < 5 * 60 * 1000;
}

function formatDate(dateStr) {
    const d = new Date(dateStr);
    return d.toLocaleString("it-IT", { dateStyle: "short", timeStyle: "short" });
}

function SensorCard({ id, type, reference, registrationTime, lastSeen }) {
    const active = isSensorActive(lastSeen);
    return (
        <div className={`sensor-card ${active ? "active-sensor" : "inactive-sensor"}`}>
            <span className="sensor-id">{id}</span>
            <span className="sensor-type">{type} ({reference})</span>
            <span className="sensor-date">
                Registrato: {formatDate(registrationTime)}
            </span>
            <span className="sensor-date">
                Ultima attività: {formatDate(lastSeen)}
            </span>
            <span className="sensor-status">
                {active ? "Attivo" : "Non attivo"}
            </span>
        </div>
    );
}

export default SensorCard;