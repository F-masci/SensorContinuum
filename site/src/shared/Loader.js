import React from "react";
import "./Loader.css";

function Loader({ text = "Caricamento..." }) {
    return (
        <div className="loader">
            <div className="loader-spinner"></div>
            <span className="loader-text">{text}</span>
        </div>
    );
}

export default Loader;