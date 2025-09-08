import React, { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import MacrozoneCard from "./MacrozoneCard";
import HubCard from "../shared/HubCard";
import Loader from '../shared/Loader'

import './Region.css';

function RegionDashboard() {
    const { name } = useParams();
    const navigate = useNavigate();
    const [region, setRegion] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const url = process.env.REACT_APP_REGION_SEARCH_NAME_URL.replace("{name}", encodeURIComponent(name));
        fetch(url)
            .then((res) => res.json())
            .then((data) => {
                setRegion(data);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    }, [name]);

    if (loading) {
        return <Loader text="Caricamento dati macrozone..." />
    }

    if (!region) {
        return <div className="dashboard"><p>Regione non trovata.</p></div>;
    }

    return (
        <div className="region-dashboard">
            <header className="region-dashboard-header">
                <div className="d-flex align-items-center">
                    <button
                        type="button"
                        className="btn btn-light me-2"
                        onClick={() => navigate(-1)}
                        title="Torna indietro"
                    >
                        <i className="bi bi-arrow-left"></i>
                    </button>
                    <h1 className="region-dashboard-title mb-0">{region.name}</h1>
                </div>
                <div className="macrozone-cards-container">
                    {region.macrozones.map((macrozone) => (
                        <MacrozoneCard
                            key={macrozone.name}
                            name={macrozone.name}
                            zoneCount={macrozone.zone_count}
                            lat={macrozone.lat}
                            lon={macrozone.lon}
                            creationTime={macrozone.creation_time}
                        />
                    ))}
                </div>
                <div className="region-hubs-container">
                    <h2 className="text-align-center" >Hubs</h2>
                    <div className="region-hubs-list grid-hubs">
                        {region.hubs.map((hub) => (
                            <HubCard
                                key={hub.id}
                                id={hub.id}
                                service={hub.service}
                                lastSeen={hub.last_seen}
                                registrationTime={hub.registration_time}
                            />
                        ))}
                    </div>
                </div>
            </header>
        </div>
    );
}

export default RegionDashboard;