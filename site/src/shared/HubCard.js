import React from "react";

import './HubCard.css';

function isHubActive(lastSeen) {
    if (!lastSeen) return false;
    const lastSeenDate = new Date(lastSeen);
    if (isNaN(lastSeenDate.getTime())) return false;
    const now = Date.now(); // UTC timestamp
    return (now - lastSeenDate.getTime()) < 5 * 60 * 1000;
}

function formatDate(dateStr) {
    const d = new Date(dateStr);
    return d.toLocaleString("it-IT", { dateStyle: "short", timeStyle: "short" });
}

function HubCard({ id, service, lastSeen, registrationTime }) {
    const active = isHubActive(lastSeen);
    return (
        <div className={`hub-card ${active ? "active-hub" : "inactive-hub"}`}>
            <span className="hub-id">{id}</span>
            <span className="hub-service">{service}</span>
            <span className="hub-date">
                Registrato: {formatDate(registrationTime)}
            </span>
            <span className="hub-date">
                Ultima attività: {formatDate(lastSeen)}
            </span>
            <span className="hub-status">
                {active ? "Attivo" : "Non attivo"}
            </span>
        </div>
    );
}

export default HubCard;