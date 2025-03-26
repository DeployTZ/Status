document.addEventListener('DOMContentLoaded', () => {
    const overallStatusEl = document.getElementById('overall-status');
    const overallStatusDot = overallStatusEl.querySelector('.status-dot');
    const overallStatusText = overallStatusEl.querySelector('.status-text');

    const currentStatusTextEl = document.getElementById('current-status-text');
    const currentResponseTimeEl = document.getElementById('current-response-time');
    const lastCheckedTimeEl = document.getElementById('last-checked-time');

    const uptime24hEl = document.getElementById('uptime-24h');
    const uptime7dEl = document.getElementById('uptime-7d');
    const uptime30dEl = document.getElementById('uptime-30d');

    const timelineContainer = document.getElementById('timeline-container');
    const currentTimeEl = document.getElementById('current-time');

    const HISTORY_DAYS = 90; // Should match Go backend const if possible

    // --- Utility Functions ---
    function formatTimeAgo(date) {
        if (!date) return 'never';
        const seconds = Math.floor((new Date() - date) / 1000);
        let interval = Math.floor(seconds / 31536000);
        if (interval > 1) return interval + " years ago";
        interval = Math.floor(seconds / 2592000);
        if (interval > 1) return interval + " months ago";
        interval = Math.floor(seconds / 86400);
        if (interval > 1) return interval + " days ago";
        interval = Math.floor(seconds / 3600);
        if (interval > 1) return interval + " hours ago";
        interval = Math.floor(seconds / 60);
        if (interval > 1) return interval + " minutes ago";
        if (seconds < 10) return "just now";
        return Math.floor(seconds) + " seconds ago";
    }

     function formatTimestamp(date) {
        if (!date) return '--';
        return date.toLocaleString(); // Adjust formatting as needed
    }

    function updateCurrentTime() {
        currentTimeEl.textContent = formatTimestamp(new Date());
    }

    // --- Data Fetching Functions ---
    async function fetchApi(url) {
        try {
            const response = await fetch(url);
            if (!response.ok) {
                console.error(`API Error (${response.status}): ${await response.text()}`);
                return null; // Or throw an error
            }
            return await response.json();
        } catch (error) {
            console.error(`Fetch Error for ${url}:`, error);
            return null; // Or throw an error
        }
    }

    async function fetchAndUpdateCurrentStatus() {
        const data = await fetchApi('/api/status/current');
        if (!data) {
            // Handle error display if needed
             updateOverallStatus(null, 0); // Indicate error state (pass isUp=null, statusCode=0)
             currentStatusTextEl.textContent = 'Error Fetching';
             currentResponseTimeEl.textContent = '-- ms';
             lastCheckedTimeEl.textContent = 'never';
             lastCheckedTimeEl.title = '';
            return;
        }

        const checkDate = data.timestamp ? new Date(data.timestamp) : null;

        // Pass status code to the update function
        updateOverallStatus(data.is_up, data.status_code);

        currentStatusTextEl.textContent = data.is_up ? 'Operational' : (data.status_code >= 400 ? `Error (${data.status_code})` : 'Down');
        currentResponseTimeEl.textContent = `${data.response_time_ms} ms`;
        lastCheckedTimeEl.textContent = formatTimeAgo(checkDate);
        lastCheckedTimeEl.title = formatTimestamp(checkDate); // Tooltip for exact time
    }

    async function fetchAndUpdateUptime() {
        const data = await fetchApi('/api/status/uptime');
         if (!data) {
             uptime24hEl.textContent = 'Error';
             uptime7dEl.textContent = 'Error';
             uptime30dEl.textContent = 'Error';
             return;
         }
        uptime24hEl.textContent = data.uptime24h ?? '--%';
        uptime7dEl.textContent = data.uptime7d ?? '--%';
        uptime30dEl.textContent = data.uptime30d ?? '--%';
    }

    async function fetchAndUpdateHistory() {
        const historyData = await fetchApi('/api/status/history');
         if (!historyData) {
             timelineContainer.innerHTML = '<div class="loading-placeholder">Error loading history.</div>';
             return;
         }

        renderHistoryTimeline(historyData);
    }

    // --- UI Update Functions ---
     function updateOverallStatus(isUp, statusCode) { // Pass statusCode for better context
        overallStatusEl.classList.remove('loading', 'operational', 'outage', 'unknown');
        // Ensure the status-badge class is always present
        overallStatusEl.classList.add('overall-status'); // Renamed from status-badge to overall-status in CSS

        if (isUp === null) { // Error during fetch or no data yet
             overallStatusEl.classList.add('unknown'); // Use 'unknown' state
             overallStatusText.textContent = 'Status Unknown';
        } else if (isUp) {
            overallStatusEl.classList.add('operational');
            overallStatusText.textContent = 'Operational'; // Simplified text
        } else {
            overallStatusEl.classList.add('outage');
             // Add more specific text if possible
            if (statusCode && statusCode > 0) {
                 overallStatusText.textContent = `Outage (${statusCode})`;
            } else {
                 overallStatusText.textContent = 'Outage'; // Generic outage/timeout
            }
        }
    }


    function renderHistoryTimeline(historyData) {
        timelineContainer.innerHTML = ''; // Clear previous bars or loading placeholder
        timelineContainer.style.setProperty('--history-days', HISTORY_DAYS);

        const now = new Date();
        const daysMap = new Map(); // Key: YYYY-MM-DD, Value: { up: count, down: count, total: count }

        // Aggregate status checks by day
        historyData.forEach(record => {
            const recordDate = new Date(record.timestamp);
            // Ensure we only consider data within the HISTORY_DAYS window relative to *today*
            const dayDiff = (now.setHours(0,0,0,0) - recordDate.setHours(0,0,0,0)) / (1000 * 60 * 60 * 24);
            if (dayDiff >= 0 && dayDiff < HISTORY_DAYS) {
                const dayKey = recordDate.toISOString().split('T')[0];
                if (!daysMap.has(dayKey)) {
                    daysMap.set(dayKey, { up: 0, down: 0, total: 0, date: recordDate });
                }
                const dayStats = daysMap.get(dayKey);
                dayStats.total++;
                if (record.is_up) {
                    dayStats.up++;
                } else {
                    dayStats.down++;
                }
            }
        });

        // Generate bars for the last HISTORY_DAYS
        for (let i = HISTORY_DAYS - 1; i >= 0; i--) {
            const date = new Date(now);
            date.setDate(now.getDate() - i);
            const dayKey = date.toISOString().split('T')[0];
            const dayStats = daysMap.get(dayKey);

            const bar = document.createElement('div');
            bar.classList.add('timeline-bar');

            let tooltipText = `${date.toLocaleDateString()} - No data`;
            let barClass = 'nodata';

            if (dayStats) {
                if (dayStats.down === 0 && dayStats.up > 0) {
                    barClass = 'operational';
                    tooltipText = `${date.toLocaleDateString()} - Operational`;
                } else if (dayStats.up === 0 && dayStats.down > 0) {
                    barClass = 'outage';
                    tooltipText = `${date.toLocaleDateString()} - Outage (${dayStats.down} checks)`;
                } else if (dayStats.up > 0 && dayStats.down > 0) {
                    // barClass = 'partial'; // Keep class if needed, but rely on gradient
                     const outagePercent = (dayStats.down / dayStats.total * 100).toFixed(1);
                    tooltipText = `${date.toLocaleDateString()} - Partial Outage (${outagePercent}% downtime)`;

                     // Style directly using CSS variables for the gradient
                     const gradientPercentage = Math.max(10, Math.min(90, (dayStats.down / dayStats.total) * 100));
                     // Use CSS variables correctly in the style string
                     bar.style.background = `linear-gradient(to top, var(--danger-color) ${gradientPercentage}%, var(--accent-color) ${gradientPercentage}%)`;
                     barClass = 'partial'; // Add class anyway for potential future styling/selection
                }
                // Add more details to tooltip if needed
                tooltipText += ` (${dayStats.up} up, ${dayStats.down} down)`;
            }

            bar.classList.add(barClass);
            bar.setAttribute('data-tooltip', tooltipText);
            timelineContainer.appendChild(bar);
        }
    }


    // --- Initial Load & Interval ---
    function initialLoad() {
        fetchAndUpdateCurrentStatus();
        fetchAndUpdateUptime();
        fetchAndUpdateHistory();
        updateCurrentTime();
    }

    initialLoad(); // Load data immediately on page load

    // Refresh current status, uptime, and clock more frequently
    setInterval(() => {
         fetchAndUpdateCurrentStatus();
         fetchAndUpdateUptime();
         updateCurrentTime();
     }, 60 * 1000); // Every 60 seconds

    // Refresh history less frequently (e.g., every 5-15 minutes)
    setInterval(fetchAndUpdateHistory, 5 * 60 * 1000); // Every 5 minutes
});
