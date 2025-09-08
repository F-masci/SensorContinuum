import React from "react";
import { Modal } from "react-bootstrap";
import DataTable from "react-data-table-component";
import Loader from "../shared/Loader";

function formatTimestamp(ts) {
    if (!ts) return "-";
    const date = new Date(ts * 1000);
    if (isNaN(date.getTime())) return "-";
    return date.toLocaleString("it-IT", {
        year: "numeric",
        month: "2-digit",
        day: "2-digit",
        hour: "2-digit",
        minute: "2-digit"
    });
}

const columns = [
    { name: "Timestamp", selector: row => formatTimestamp(row.timestamp), sortable: true },
    { name: "Tipo", selector: row => row.type, sortable: true },
    { name: "Min", selector: row => row.min, sortable: true },
    { name: "Max", selector: row => row.max, sortable: true },
    { name: "Media", selector: row => row.avg, sortable: true }
];

function ZoneAggregateDataModal({ show, onHide, zone, data, loading }) {
    return (
        <Modal show={show} onHide={onHide} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>Dati aggregati zona {zone?.name}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {loading ? (
                    <Loader text="Caricamento dati zona..." />
                ) : (
                    <DataTable
                        columns={columns}
                        data={data}
                        pagination
                        striped
                        highlightOnHover
                        dense
                        noDataComponent="Nessun dato aggregato"
                    />
                )}
            </Modal.Body>
        </Modal>
    );
}

export default ZoneAggregateDataModal;