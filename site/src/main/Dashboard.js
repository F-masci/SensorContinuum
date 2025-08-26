import React, { useEffect, useState } from "react";
import RegionCard from "./RegionCard";

import './Main.css';

function MainDashboard() {
    const [regions, setRegions] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch(process.env.REACT_APP_REGION_LIST_URL)
            .then((res) => res.json())
            .then((data) => {
                setRegions(data);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    }, []);

    return (
        <div className="main-dashboard">
            <header className="main-dashboard-header">
                <h1 className="main-dashboard-title">Sensor Continuum</h1>
            </header>
            {loading ? (
                <p>Caricamento...</p>
            ) : (
                <div className="region-cards-container">
                    {regions.map((region) => (
                        <RegionCard
                            key={region.name}
                            name={region.name}
                            macrozoneCount={region.macrozone_count}
                        />
                    ))}
                </div>
            )}
        </div>
    );
}

export default MainDashboard;
