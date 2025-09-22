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

    const [macrozoneTrend, setMacrozoneTrend] = useState({});
    const [macrozoneTrendLoading, setMacrozoneTrendLoading] = useState(false);

    // Stato per ricerca dei trend
    const prevYesterday = new Date();
    prevYesterday.setDate(prevYesterday.getDate() - 2);
    const yyyy = prevYesterday.getFullYear();
    const mm = String(prevYesterday.getMonth() + 1).padStart(2, "0");
    const dd = String(prevYesterday.getDate()).padStart(2, "0");
    const defaultDate = `${yyyy}-${mm}-${dd}`;

    // Stati per trend
    const [trendDate, setTrendDate] = useState(defaultDate);
    const [trendDays, setTrendDays] = useState(60);
    const [searchingTrend, setSearchingTrend] = useState(false);

    // Stato per variazioni annuali
    const [macrozoneVariation, setMacrozoneVariation] = useState({});
    const [macrozoneVariationLoading, setMacrozoneVariationLoading] = useState(false);
    const [variationDate, setVariationDate] = useState(defaultDate); // default = due giorni fa

    // Stato per correlazioni tra variazioni
    const [macrozoneVariationCorrelation, setMacrozoneVariationCorrelation] = useState({});
    const [macrozoneVariationCorrelationLoading, setMacrozoneVariationCorrelationLoading] = useState(false);
    const [correlationRadius, setCorrelationRadius] = useState(10000000); // default radius



    // Fetch dati regione
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

    // Funzione per fetch delle variazioni annuali
    const fetchMacrozoneVariation = () => {
        if (!region) return;

        setMacrozoneVariationLoading(true);

        const params = new URLSearchParams();
        if (variationDate) params.append("date", variationDate);

        const url = `${process.env.REACT_APP_MACROZONE_DATA_VARIATION_URL.replace("{region}", encodeURIComponent(region.name))}?${params.toString()}`;

        fetch(url)
            .then(res => res.json())
            .then(data => setMacrozoneVariation(data))
            .catch(() => setMacrozoneVariation({}))
            .finally(() => setMacrozoneVariationLoading(false));
    };

    // Funzione per fetch delle correlazioni tra variazioni
    const fetchVariationCorrelation = () => {
        if (!region) return;

        setMacrozoneVariationCorrelationLoading(true);

        const params = new URLSearchParams();
        if (correlationRadius) params.append("radius", correlationRadius);

        const url = `${process.env.REACT_APP_MACROZONE_DATA_VARIATION_CORRELATION_URL.replace("{region}", encodeURIComponent(region.name))}?${params.toString()}`;

        fetch(url)
            .then(res => res.json())
            .then(data => setMacrozoneVariationCorrelation(data))
            .catch(() => setMacrozoneVariationCorrelation({}))
            .finally(() => setMacrozoneVariationCorrelationLoading(false));
    };

    // Fetch dati trend macrozone
    const fetchMacrozoneTrend = () => {
        if (!region) return;
        setMacrozoneTrendLoading(true);
        setSearchingTrend(true);

        // Costruzione query string
        const params = new URLSearchParams();
        if (trendDays) params.append("days", trendDays);
        if (trendDate) params.append("date", trendDate);

        // Esempio: /macrozone/data/trend/{region}?days=7&date=2025-09-22
        const url = `${process.env.REACT_APP_MACROZONE_DATA_TREND_URL.replace("{region}", encodeURIComponent(region.name))}?${params.toString()}`;

        fetch(url)
            .then(res => res.json())
            .then(data => setMacrozoneTrend(data))
            .catch(() => setMacrozoneTrend({}))
            .finally(() => {
                setMacrozoneTrendLoading(false);
                setSearchingTrend(false);
            });
    };


    if (loading) {
        return <Loader text="Caricamento dati macrozone..." />
    }

    if (!region) {
        return <div className="dashboard"><p>Regione non trovata.</p></div>;
    }

    const renderVariationData = () => {
        if (!macrozoneVariation || Object.keys(macrozoneVariation).length === 0) {
            return <p className="trend-no-data">Nessun dato disponibile.</p>;
        }

        return (
            <div className="variation-macrozone-container">
                {Object.entries(macrozoneVariation).map(([macrozoneName, types]) => (
                    <div key={macrozoneName} className="variation-macrozone-card">
                        <h2 className="variation-macrozone-title">{macrozoneName}</h2>
                        <div className="variation-macrozone-types">
                            {Object.entries(types).map(([type, data]) => {
                                const deltaPerc = data.delta_perc != null ? data.delta_perc.toFixed(2) : "N/A";
                                let deltaIcon = "bi-dash";
                                let deltaColor = "#215b93";
                                if (data.delta_perc > 0) { deltaIcon = "bi-arrow-up"; deltaColor = "#28a745"; }
                                else if (data.delta_perc < 0) { deltaIcon = "bi-arrow-down"; deltaColor = "#dc3545"; }

                                return (
                                    <div key={type} className="trend-type-card">
                                        <h4 className="trend-type-title">{type}</h4>
                                        <p><strong>Macrozone:</strong> {data.macrozone || "N/A"}</p>
                                        <p><strong>Current:</strong> {data.current?.toFixed(2)}</p>
                                        <p><strong>Previous:</strong> {data.previous?.toFixed(2)}</p>
                                        <p><strong>Delta %:</strong> {deltaPerc} <i className={`bi ${deltaIcon}`} style={{color: deltaColor}}></i></p>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </div>
        );
    };

    const renderVariationCorrelationData = () => {
        if (!macrozoneVariationCorrelation || Object.keys(macrozoneVariationCorrelation).length === 0) {
            return <p className="trend-no-data">Nessun dato di correlazione disponibile.</p>;
        }

        return (
            <div className="variation-macrozone-container">
                {Object.entries(macrozoneVariationCorrelation).map(([macrozoneName, types]) => (
                    <div key={macrozoneName} className="variation-macrozone-card">
                        <h2 className="variation-macrozone-title">{macrozoneName}</h2>
                        <div className="variation-macrozone-types">
                            {Object.entries(types).map(([type, data]) => {
                                const absError = data.abs_error != null ? data.abs_error.toFixed(2) : "N/A";
                                const zScore = data.z_score != null ? data.z_score.toFixed(2) : "N/A";
                                const neighborMean = data.neighbor_mean != null ? data.neighbor_mean.toFixed(2) : "N/A";
                                const neighborStd = data.neighbor_std_dev != null ? data.neighbor_std_dev.toFixed(2) : "N/A";

                                return (
                                    <div key={type} className="trend-type-card">
                                        <h4 className="trend-type-title">{type}</h4>
                                        <p><strong>Macrozone:</strong> {data.macrozone || "N/A"}</p>
                                        <p><strong>Current Variation:</strong> {data.variation?.current?.toFixed(2) || "N/A"}</p>
                                        <p><strong>Previous Variation:</strong> {data.variation?.previous?.toFixed(2) || "N/A"}</p>
                                        <p><strong>Neighbor Mean:</strong> {neighborMean}</p>
                                        <p><strong>Neighbor Std Dev:</strong> {neighborStd}</p>
                                        <p><strong>Abs Error:</strong> {absError}</p>
                                        <p><strong>Z-score:</strong> {zScore}</p>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </div>
        );
    };

    const renderTrendData = () => {
        if (!macrozoneTrend || Object.keys(macrozoneTrend).length === 0) {
            return <p className="trend-no-data">Nessun dato di trend disponibile.</p>;
        }

        return (
            <div className="trend-macrozone-container">
                {Object.entries(macrozoneTrend).map(([macrozoneName, types]) => (
                    <div key={macrozoneName} className="trend-macrozone-card">
                        <h2 className="trend-macrozone-title">{macrozoneName}</h2>
                        <div className="trend-macrozone-types">
                            {Object.entries(types).map(([type, data]) => {
                                const correlation = data.correlation != null ? data.correlation.toFixed(3) : "N/A";
                                const slopeThreshold = 1e-6;
                                const divergenceThreshold = 0.2;

                                // Slope Macro
                                const slopeMacroVal = data.slope_macro;
                                const slopeMacroStr = slopeMacroVal != null ? slopeMacroVal.toExponential(3) : "N/A";
                                let slopeMacroIcon = "bi-dash";
                                let slopeMacroColor = "#215b93";
                                if (slopeMacroVal > slopeThreshold) { slopeMacroIcon = "bi-arrow-up"; slopeMacroColor = "#28a745"; }
                                else if (slopeMacroVal < -slopeThreshold) { slopeMacroIcon = "bi-arrow-down"; slopeMacroColor = "#dc3545"; }

                                // Slope Region
                                const slopeRegionVal = data.slope_region;
                                const slopeRegionStr = slopeRegionVal != null ? slopeRegionVal.toExponential(3) : "N/A";
                                let slopeRegionIcon = "bi-dash";
                                let slopeRegionColor = "#215b93";
                                if (slopeRegionVal > slopeThreshold) { slopeRegionIcon = "bi-arrow-up"; slopeRegionColor = "#28a745"; }
                                else if (slopeRegionVal < -slopeThreshold) { slopeRegionIcon = "bi-arrow-down"; slopeRegionColor = "#dc3545"; }

                                // Divergence
                                const divergenceVal = data.divergence;
                                const divergenceStr = divergenceVal != null ? divergenceVal.toExponential(3) : "N/A";
                                let divergenceIcon = "bi-dash";
                                let divergenceColor = "#28a745"; // verde se bassa divergenza
                                if (divergenceVal != null && divergenceVal > divergenceThreshold) {
                                    divergenceIcon = "bi-exclamation-circle";
                                    divergenceColor = "#FFC107"; // giallo se divergente
                                }

                                return (
                                    <div key={type} className="trend-type-card">
                                        <h4 className="trend-type-title">{type}</h4>
                                        <p><strong>Macrozone:</strong> {data.macrozone || "N/A"}</p>
                                        <p><strong>Correlation:</strong> {correlation}</p>
                                        <p><strong>Slope Macro:</strong> {slopeMacroStr} <i className={`bi ${slopeMacroIcon}`} style={{color: slopeMacroColor}}></i></p>
                                        <p><strong>Slope Region:</strong> {slopeRegionStr} <i className={`bi ${slopeRegionIcon}`} style={{color: slopeRegionColor}}></i></p>
                                        <p><strong>Divergence:</strong> {divergenceStr} <i className={`bi ${divergenceIcon}`} style={{color: divergenceColor}}></i></p>
                                    </div>
                                );
                            })}
                        </div>
                    </div>
                ))}
            </div>
        );
    };


    return (
        <div className="region-dashboard">
            <header className="region-dashboard-header">
                <div className="d-flex align-items-center mb-4">
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

                <div className="macrozone-cards-container mb-4">
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

                <div className="region-hubs-container mb-4">
                    <h2>Hubs</h2>
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

                {/* Form di ricerca variazioni */}
                <form className="search-form" onSubmit={(e) => { e.preventDefault(); fetchMacrozoneVariation(); }}>
                    <h3>Variazioni Annuali Macrozone</h3>
                    <div className="search-fields">
                        <label>
                            Giorno:
                            <input
                                type="date"
                                value={variationDate}
                                onChange={e => setVariationDate(e.target.value)}
                                className="form-control"
                                max={defaultDate} // fino a ieri
                            />
                        </label>
                    </div>
                    <button type="submit" disabled={macrozoneVariationLoading}>
                        {macrozoneVariationLoading ? "Caricamento..." : "Cerca"}
                    </button>
                </form>

                <div className="mb-4">
                    {macrozoneVariationLoading ? <Loader text="Caricamento variazioni..." /> : renderVariationData()}
                </div>

                {/* Form di ricerca correlazione variazione */}
                <form className="search-form" onSubmit={(e) => { e.preventDefault(); fetchVariationCorrelation(); }}>
                    <h3>Correlazione Variazioni Annuali</h3>
                    <div className="search-fields">
                        <label>
                            Raggio:
                            <input
                                type="number"
                                value={correlationRadius}
                                onChange={e => setCorrelationRadius(e.target.value)}
                                className="form-control"
                                min={1}
                            />
                        </label>
                    </div>
                    <button type="submit" disabled={macrozoneVariationCorrelationLoading}>
                        {macrozoneVariationCorrelationLoading ? "Calcolo in corso..." : "Calcola Correlazione"}
                    </button>
                </form>

                <div className="mb-4">
                    {macrozoneVariationCorrelationLoading ? <Loader text="Calcolo correlazioni..." /> : renderVariationCorrelationData()}
                </div>

                {/* Form di ricerca trend */}
                <form className="search-form" onSubmit={(e) => { e.preventDefault(); fetchMacrozoneTrend(); }}>
                    <h3>Trend Macrozone</h3>
                    <div className="search-fields">
                        <label>
                            Giorni da considerare:
                            <input
                                type="number"
                                value={trendDays}
                                onChange={e => setTrendDays(e.target.value)}
                                className="form-control"
                                min={1}
                            />
                        </label>
                        <label>
                            Giorno:
                            <input
                                type="date"
                                value={trendDate}
                                onChange={e => setTrendDate(e.target.value)}
                                className="form-control"
                                max={defaultDate} // defaultDate = ieri
                            />
                        </label>
                    </div>
                    <button type="submit" disabled={searchingTrend}>
                        {searchingTrend ? "Ricerca in corso..." : "Cerca trend"}
                    </button>
                </form>

                <div className="mb-4">
                    {macrozoneTrendLoading ? <Loader text="Caricamento trend..." /> : renderTrendData()}
                </div>
            </header>
        </div>
    );
}

export default RegionDashboard;
