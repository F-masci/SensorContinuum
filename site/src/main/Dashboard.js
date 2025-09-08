import React, { useEffect, useState } from "react";
import RegionCard from "./RegionCard";
import Loader from '../shared/Loader'
import './Main.css';

function MainDashboard() {
    const [regions, setRegions] = useState([]);
    const [loading, setLoading] = useState(true);

    // Stato per la ricerca
    const [searchLat, setSearchLat] = useState("");
    const [searchLon, setSearchLon] = useState("");
    const [searchRadius, setSearchRadius] = useState("");
    const [searchResults, setSearchResults] = useState([]);
    const [searchLoading, setSearchLoading] = useState(false);

    useEffect(() => {
        fetch(process.env.REACT_APP_REGION_LIST_URL)
            .then((res) => res.json())
            .then((data) => {
                setRegions(data);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    }, []);

    const handleSearch = (e) => {
        e.preventDefault();
        setSearchLoading(true);
        setSearchResults([]);
        fetch(`${process.env.REACT_APP_MACROZONE_DATA_AGGREGATED_LOCATION_URL}?lat=${searchLat}&lon=${searchLon}&radius=${searchRadius}`)
            .then((res) => res.json())
            .then((data) => {
                setSearchResults(data);
                setSearchLoading(false);
            })
            .catch(() => setSearchLoading(false));
    };

    return (
        <div className="main-dashboard">
            <header className="main-dashboard-header">
                <h1 className="main-dashboard-title">Sensor Continuum</h1>
            </header>

            {loading ? (
                <Loader text="Caricamento dati regioni..." />
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

            {/* Form di ricerca sotto le card */}
            <form className="search-form" onSubmit={handleSearch}>
                <div className="search-fields">
                    <label>
                        Latitudine:
                        <input
                            type="text"
                            value={searchLat}
                            onChange={(e) => setSearchLat(e.target.value)}
                            required
                        />
                    </label>
                    <label>
                        Longitudine:
                        <input
                            type="text"
                            value={searchLon}
                            onChange={(e) => setSearchLon(e.target.value)}
                            required
                        />
                    </label>
                    <label>
                        Raggio (m):
                        <input
                            type="text"
                            value={searchRadius}
                            onChange={(e) => setSearchRadius(e.target.value)}
                            required
                        />
                    </label>
                </div>
                <button type="submit" disabled={searchLoading}>Cerca valori</button>
            </form>

            {searchLoading && <p className="search-loading">Ricerca in corso...</p>}
            {searchResults.length > 0 && (
                <div className="search-results-card">
                    <h2>Valori aggregati rilevati</h2>
                    <ul>
                        {[...searchResults]
                            .sort((a, b) => a.type.localeCompare(b.type))
                            .map((val, idx) => (
                                <li key={idx}>
                                    <div>
                                        <strong>Tipo:</strong> {val.type} <br />
                                        <strong>Min:</strong> {val.min?.toFixed(2)} &nbsp;
                                        <strong>Max:</strong> {val.max?.toFixed(2)} <br />
                                        <strong>Avg:</strong> {val.avg?.toFixed(2)} &nbsp;
                                        <strong>Weighted Avg:</strong> {val.weighted_avg?.toFixed(2)} <br />
                                        <strong>Timestamp:</strong> {new Date(val.timestamp * 1000).toLocaleString()}
                                    </div>
                                </li>
                            ))
                        }
                    </ul>
                </div>
            )}
        </div>
    );
}

export default MainDashboard;