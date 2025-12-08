const API_BASE_URL = 'http://localhost:8080/api/v1';

// DOM elements
const elements = {
    loading: document.getElementById('loading'),
    error: document.getElementById('error'),
    dashboard: document.getElementById('dashboard'),
    refreshBtn: document.getElementById('refreshBtn'),
    location: document.getElementById('location'),
    timeRange: document.getElementById('timeRange'),
    lastUpdate: document.getElementById('lastUpdate'),
};

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadLatestSnapshot();
    
    elements.refreshBtn.addEventListener('click', loadLatestSnapshot);
    
    // Auto-refresh every 60 seconds
    setInterval(loadLatestSnapshot, 60000);
});

// Load latest snapshot from API
async function loadLatestSnapshot() {
    try {
        showLoading();
        hideError();
        
        const location = elements.location.value || 'Los Angeles';
        const response = await fetch(`${API_BASE_URL}/snapshots/latest?location=${encodeURIComponent(location)}`);
        
        if (!response.ok) {
            throw new Error(`API error: ${response.status} ${response.statusText}`);
        }
        
        const snapshot = await response.json();
        updateDashboard(snapshot);
        updateLastUpdateTime();
        hideLoading();
        
    } catch (error) {
        console.error('Error loading snapshot:', error);
        showError(`Failed to load data: ${error.message}`);
        hideLoading();
    }
}

// Update dashboard with snapshot data
function updateDashboard(snapshot) {
    // Weather
    updateElement('temp', snapshot.weather?.temperature_c, 1);
    updateElement('humidity', snapshot.weather?.humidity, 0);
    updateElement('wind', snapshot.weather?.wind_speed_ms, 1);
    updateElement('clouds', snapshot.weather?.cloud_cover, 0);
    
    // Environment (Air Quality)
    updateElement('pm25', snapshot.environment?.pm25, 1);
    updateElement('pm10', snapshot.environment?.pm10, 1);
    updateElement('ozone', snapshot.environment?.ozone, 1);
    updateElement('no2', snapshot.environment?.no2, 1);
    
    // Energy
    updateElement('gridLoad', snapshot.energy?.grid_load, 0);
    updateElement('renewable', snapshot.energy?.renewable_percent, 1);
    updateElement('carbon', snapshot.energy?.carbon_intensity_gco2_kwh, 1);
    updateElement('gridUtil', snapshot.energy?.grid_utilization_percent, 1);
    
    // Finance
    updateElement('nasdaq', snapshot.finance?.nasdaq_index, 2);
    updateElement('stock', snapshot.finance?.stock_price, 2);
    updateElement('volume', snapshot.finance?.volume_traded ? (snapshot.finance.volume_traded / 1000000).toFixed(1) : '--', null);
    
    if (snapshot.finance?.stock_symbol) {
        document.getElementById('stockSymbol').textContent = snapshot.finance.stock_symbol;
    }
    
    // Health
    updateElement('fluCases', snapshot.health?.flu_cases, 0);
    updateElement('ili', snapshot.health?.ili_percent, 2);
    updateElement('hospital', snapshot.health?.hospital_admissions, 0);
    
    // Agriculture
    document.getElementById('cropType').textContent = snapshot.agriculture?.crop_type || '--';
    updateElement('yield', snapshot.agriculture?.crop_yield, 1);
    updateElement('cropPrice', snapshot.agriculture?.price_per_bushel, 2);
    updateElement('production', snapshot.agriculture?.production_bushels ? (snapshot.agriculture.production_bushels / 1000000000).toFixed(1) : '--', null);
    
    // Disasters
    updateElement('disasters', snapshot.disasters?.active_disasters, 0);
    document.getElementById('disasterType').textContent = snapshot.disasters?.disaster_type || '--';
    updateElement('severity', snapshot.disasters?.severity, 1);
    updateElement('counties', snapshot.disasters?.affected_counties, 0);
    
    // Mobility (Wildlife)
    updateElement('species', snapshot.mobility?.active_species, 0);
    updateElement('animals', snapshot.mobility?.animals_tracked, 0);
    updateElement('migrationPace', snapshot.mobility?.avg_migration_pace_km_day, 1);
}

// Helper function to update element with formatted value
function updateElement(id, value, decimals) {
    const element = document.getElementById(id);
    if (!element) return;
    
    if (value === null || value === undefined || value === 0) {
        element.textContent = '--';
        return;
    }
    
    if (decimals !== null) {
        element.textContent = typeof value === 'number' ? value.toFixed(decimals) : value;
    } else {
        element.textContent = value;
    }
}

// Update last update timestamp
function updateLastUpdateTime() {
    const now = new Date();
    elements.lastUpdate.textContent = `Last updated: ${now.toLocaleTimeString()}`;
}

// UI helper functions
function showLoading() {
    elements.loading.style.display = 'block';
    elements.dashboard.style.opacity = '0.5';
}

function hideLoading() {
    elements.loading.style.display = 'none';
    elements.dashboard.style.opacity = '1';
}

function showError(message) {
    elements.error.textContent = message;
    elements.error.style.display = 'block';
}

function hideError() {
    elements.error.style.display = 'none';
}

// Export for debugging
window.edgeSight = {
    loadLatestSnapshot,
    API_BASE_URL
};
