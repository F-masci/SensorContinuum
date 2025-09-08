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
    { name: "Valore", selector: row => row.data, sortable: true }
];

function SensorDataModal({ show, onHide, sensor, data, loading }) {
    return (
        <Modal show={show} onHide={onHide} size="lg" centered>
            <Modal.Header closeButton>
                <Modal.Title>Rilevazioni sensore {sensor?.id}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {loading ? (
                    <Loader text="Caricamento dati sensore..." />
                ) : (
                    <DataTable
                        columns={columns}
                        data={data}
                        pagination
                        striped
                        highlightOnHover
                        dense
                        noDataComponent="Nessuna rilevazione"
                    />
                )}
            </Modal.Body>
        </Modal>
    );
}

export default SensorDataModal;