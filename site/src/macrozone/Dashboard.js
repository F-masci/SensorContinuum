import React, { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { useNavigate } from "react-router-dom";
import DataTable from "react-data-table-component";
import { Card, Badge, Button } from "react-bootstrap";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap-icons/font/bootstrap-icons.css";

import "./Macrozone.css";
import SensorDataModal from "./SensorDataModal";
import ZoneAggregateDataModal from "./ZoneAggregateDataModal";
import MacrozoneAggregateDataModal from "./MacrozoneAggregateDataModal";
import Loader from "../shared/Loader";

function isActive(lastSeen) {
    if (!lastSeen) return false;
    const lastSeenDate = new Date(lastSeen);
    if (isNaN(lastSeenDate.getTime())) return false;
    const now = Date.now();
    return (now - lastSeenDate.getTime()) < 10 * 60 * 1000;
}

const getBadge = lastSeen =>
    isActive(lastSeen)
        ? <Badge bg="success">Attivo</Badge>
        : <Badge bg="danger">Non attivo</Badge>;

function formatDate(dateStr) {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return "-";
    return date.toLocaleString("it-IT", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit"
    });
}

function Dashboard() {
    const { regionName, macrozoneName } = useParams();
    const [macrozone, setMacrozone] = useState(null);
    const [loading, setLoading] = useState(true);
    const [searchHubs, setSearchHubs] = useState("");
    const [searchSensors, setSearchSensors] = useState({});
    const [searchZoneHubs, setSearchZoneHubs] = useState({});
    const [collapsedZones, setCollapsedZones] = useState({});
    const [showModal, setShowModal] = useState(false);
    const [selectedSensor, setSelectedSensor] = useState(null);
    const [sensorData, setSensorData] = useState([]);
    const [showZoneModal, setShowZoneModal] = useState(false);
    const [zoneData, setZoneData] = useState([]);
    const [selectedZone, setSelectedZone] = useState(null);
    const [showMacrozoneModal, setShowMacrozoneModal] = useState(false);
    const [macrozoneData, setMacrozoneData] = useState([]);
    const [sensorDataLoading, setSensorDataLoading] = useState(false);
    const [zoneDataLoading, setZoneDataLoading] = useState(false);
    const [macrozoneDataLoading, setMacrozoneDataLoading] = useState(false);

    const navigate = useNavigate();

    useEffect(() => {
        const url = process.env.REACT_APP_MACROZONE_SEARCH_NAME_URL
            .replace("{region}", encodeURIComponent(regionName))
            .replace("{name}", encodeURIComponent(macrozoneName));
        fetch(url)
            .then((res) => res.json())
            .then((data) => {
                setMacrozone(data);
                setLoading(false);
            })
            .catch(() => setLoading(false));
    }, [regionName, macrozoneName]);

    const handleShowSensorData = (sensor) => {
        setSelectedSensor(sensor);
        setShowModal(true);
        setSensorDataLoading(true);
        const url = process.env.REACT_APP_ZONE_SENSOR_DATA_RAW_URL
            .replace("{region}", encodeURIComponent(regionName))
            .replace("{macrozone}", encodeURIComponent(macrozoneName))
            .replace("{zone}", encodeURIComponent(sensor.zone_name))
            .replace("{sensor}", encodeURIComponent(sensor.id));
        fetch(url)
            .then(res => res.json())
            .then(data => setSensorData(data))
            .catch(() => setSensorData([]))
            .finally(() => setSensorDataLoading(false));
    };

    const handleCloseModal = () => {
        setShowModal(false);
        setSelectedSensor(null);
        setSensorData([]);
    };

    const handleShowZoneData = (zone) => {
        setSelectedZone(zone);
        setShowZoneModal(true);
        setZoneDataLoading(true);
        const url = process.env.REACT_APP_ZONE_DATA_AGGREGATED_URL
            .replace("{region}", encodeURIComponent(regionName))
            .replace("{macrozone}", encodeURIComponent(macrozoneName))
            .replace("{zone}", encodeURIComponent(zone.name));
        fetch(url)
            .then(res => res.json())
            .then(data => setZoneData(data))
            .catch(() => setZoneData([]))
            .finally(() => setZoneDataLoading(false));
    };

    const handleCloseZoneModal = () => {
        setShowZoneModal(false);
        setZoneData([]);
        setSelectedZone(null);
    };

    const handleShowMacrozoneData = () => {
        setShowMacrozoneModal(true);
        setMacrozoneDataLoading(true);
        const url = process.env.REACT_APP_MACROZONE_DATA_AGGREGATED_NAME_URL
            .replace("{region}", encodeURIComponent(regionName))
            .replace("{macrozone}", encodeURIComponent(macrozoneName));
        fetch(url)
            .then(res => res.json())
            .then(data => setMacrozoneData(data))
            .catch(() => setMacrozoneData([]))
            .finally(() => setMacrozoneDataLoading(false));
    };

    const handleCloseMacrozoneModal = () => {
        setShowMacrozoneModal(false);
        setMacrozoneData([]);
    };

    if (loading) {
        return <Loader text="Caricamento dati zone..." />
    }

    if (!macrozone) {
        return <div className="dashboard"><p>Macrozone non trovata.</p></div>;
    }

    const hubColumns = [
        { name: "ID", selector: row => row.id, sortable: true },
        { name: "Servizio", selector: row => row.service, sortable: true },
        { name: "Registrazione", selector: row => formatDate(row.registration_time), sortable: true },
        { name: "Ultima attività", selector: row => formatDate(row.last_seen), sortable: true },
        {
            name: "Stato",
            cell: row => getBadge(row.last_seen)
        }
    ];

    // Colonne Sensori con pulsante occhio
    const sensorColumns = [
        { name: "ID", selector: row => row.id, sortable: true },
        { name: "Tipo", selector: row => row.type, sortable: true },
        { name: "Riferimento", selector: row => row.reference, sortable: true },
        { name: "Registrazione", selector: row => formatDate(row.registration_time), sortable: true },
        { name: "Ultima attività", selector: row => formatDate(row.last_seen), sortable: true },
        {
            name: "Stato",
            cell: row => getBadge(row.last_seen)
        },
        {
            name: "",
            cell: row => (
                <Button
                    variant="outline-primary"
                    size="sm"
                    onClick={() => handleShowSensorData(row)}
                    title="Visualizza rilevazioni"
                >
                    <i className="bi bi-eye"></i>
                </Button>
            ),
            ignoreRowClick: true,
            allowOverflow: true,
            button: true
        }
    ];

    const filteredHubs = (macrozone.hubs || []).filter(
        h => Object.values(h).join(" ").toLowerCase().includes(searchHubs.toLowerCase())
    );

    return (
        <div className="macrozone-dashboard">
            <div className="card mt-4 mb-4 shadow w-100">
                <div className="card-header bg-primary text-white">
                    <div className="d-flex justify-content-between align-items-center">
                        <Button
                            variant="light"
                            size="sm"
                            className="me-2"
                            onClick={() => navigate(-1)}
                            title="Torna indietro"
                        >
                            <i className="bi bi-arrow-left"></i>
                        </Button>
                        <h1 className="mb-0"><b>Macrozona: {macrozone.name}</b></h1>
                        <Button
                            variant="light"
                            size="sm"
                            className="ms-2"
                            onClick={handleShowMacrozoneData}
                            title="Visualizza dati aggregati macrozona"
                        >
                            <i className="bi bi-eye"></i>
                        </Button>
                    </div>
                </div>
                <div className="card-body">
                    <h2 className="text-info">Hubs Macrozona</h2>
                    <DataTable
                        columns={hubColumns}
                        data={filteredHubs}
                        pagination
                        striped
                        highlightOnHover
                        dense
                        subHeader
                        subHeaderComponent={
                            <input
                                type="text"
                                placeholder="Cerca..."
                                className="form-control w-25"
                                value={searchHubs}
                                onChange={e => setSearchHubs(e.target.value)}
                            />
                        }
                    />
                </div>
            </div>
            {(macrozone.zones || []).map(zone => {
                const zoneKey = zone.name;
                const zoneHubs = macrozone.zone_hubs?.filter(h => h.zone_name === zone.name) || [];
                const zoneSensors = macrozone.sensors?.filter(s => s.zone_name === zone.name) || [];
                const filteredZoneHubs = zoneHubs.filter(
                    h => Object.values(h).join(" ").toLowerCase().includes((searchZoneHubs[zoneKey] || "").toLowerCase())
                );
                const filteredZoneSensors = zoneSensors.filter(
                    s => Object.values(s).join(" ").toLowerCase().includes((searchSensors[zoneKey] || "").toLowerCase())
                );
                const isCollapsed = collapsedZones[zoneKey] ?? true;

                return (
                    <Card key={zoneKey} className="mb-4 w-100 border-info">
                        <Card.Header className="bg-info text-white">
                            <div className="d-flex justify-content-between align-items-center">
                                <span><b>Zona: {zone.name}</b></span>
                                <div className="d-flex ms-auto">
                                    <Button
                                        variant="light"
                                        size="sm"
                                        className="me-2"
                                        onClick={() => handleShowZoneData(zone)}
                                        title="Visualizza dati aggregati zona"
                                    >
                                        <i className="bi bi-eye"></i>
                                    </Button>
                                    <Button
                                        variant="light"
                                        size="sm"
                                        onClick={() =>
                                            setCollapsedZones(prev => ({
                                                ...prev,
                                                [zoneKey]: !isCollapsed
                                            }))
                                        }
                                    >
                                        <i className={`bi ${isCollapsed ? "bi-plus" : "bi-dash"}`}></i>
                                    </Button>
                                </div>
                            </div>
                        </Card.Header>
                        {!isCollapsed && (
                            <Card.Body>
                                <h3 className="mt-2">Hubs:</h3>
                                <DataTable
                                    columns={hubColumns}
                                    data={filteredZoneHubs}
                                    pagination
                                    striped
                                    highlightOnHover
                                    dense
                                    subHeader
                                    subHeaderComponent={
                                        <input
                                            type="text"
                                            placeholder="Cerca..."
                                            className="form-control w-25"
                                            value={searchZoneHubs[zoneKey] || ""}
                                            onChange={e =>
                                                setSearchZoneHubs(prev => ({
                                                    ...prev,
                                                    [zoneKey]: e.target.value
                                                }))
                                            }
                                        />
                                    }
                                />
                                <h3 className="mt-4">Sensori:</h3>
                                <DataTable
                                    columns={sensorColumns}
                                    data={filteredZoneSensors}
                                    pagination
                                    striped
                                    highlightOnHover
                                    dense
                                    subHeader
                                    subHeaderComponent={
                                        <input
                                            type="text"
                                            placeholder="Cerca..."
                                            className="form-control w-25"
                                            value={searchSensors[zoneKey] || ""}
                                            onChange={e =>
                                                setSearchSensors(prev => ({
                                                    ...prev,
                                                    [zoneKey]: e.target.value
                                                }))
                                            }
                                        />
                                    }
                                />
                            </Card.Body>
                        )}
                    </Card>
                );
            })}
            <MacrozoneAggregateDataModal
                show={showMacrozoneModal}
                onHide={handleCloseMacrozoneModal}
                macrozone={macrozone}
                data={macrozoneData}
                loading={macrozoneDataLoading}
            ></MacrozoneAggregateDataModal>

            <ZoneAggregateDataModal
                show={showZoneModal}
                onHide={handleCloseZoneModal}
                zone={selectedZone}
                data={zoneData}
                loading={zoneDataLoading}
            ></ZoneAggregateDataModal>

            <SensorDataModal
                show={showModal}
                onHide={handleCloseModal}
                sensor={selectedSensor}
                data={sensorData}
                loading={sensorDataLoading}
            ></SensorDataModal>
        </div>
    );
}

export default Dashboard;