import './App.css';

import { BrowserRouter, Routes, Route } from "react-router-dom";
import MainDashboard from "./main/Dashboard";
import RegionDashboard from "./region/Dashboard";
import MacrozoneDashboard from "./macrozone/Dashboard";

function App() {
    return (
        <BrowserRouter>
            <div className="app">
                <Routes>
                    <Route path="/" element={<MainDashboard />} />
                    <Route path="/region/:name" element={<RegionDashboard />} />
                    <Route path="/macrozone/:regionName/:macrozoneName" element={<MacrozoneDashboard />} />
                </Routes>
            </div>
        </BrowserRouter>
    );
}

export default App;