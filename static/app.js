/**
 * TTS SPA Application
 * Single Page Application with routing, transitions, and all page functionality
 */

// =============================================
// App State
// =============================================

const AppState = {
    currentPage: null,
    data: {
        voices: [],
        effects: [],
        modifiers: ['reverb'],
        tags: ['laughter', 'laughs', 'sad', 'sigh', 'cries', 'screams', 'gasps', 'groans', 'sniffs']
    },
    dataLoaded: false,
    chart: null,
    currentAudio: null,
    currentButton: null,
    previewVolume: 0.5
};

// =============================================
// Initialization
// =============================================

document.addEventListener('DOMContentLoaded', async () => {
    // Load data
    await loadAppData();

    // Initialize routing
    initRouter();

    // Initialize page-specific functionality
    initCreatePage();
    initChartPage();

    // Navigate to initial page based on URL
    const path = window.location.pathname;
    navigateTo(path, false);
});

// =============================================
// Data Loading
// =============================================

async function loadAppData() {
    try {
        // Load voices
        const voicesRes = await fetch('/api/voices');
        if (voicesRes.ok) {
            AppState.data.voices = await voicesRes.json();
        }

        // Load effects
        const effectsRes = await fetch('/api/effects');
        if (effectsRes.ok) {
            AppState.data.effects = await effectsRes.json();
        }

        AppState.dataLoaded = true;

        // Populate UI elements
        populateChips();
        populateAudioLists();
    } catch (err) {
        console.error('Error loading app data:', err);
    }
}

function populateChips() {
    const voicesGrid = document.getElementById('voices-chips');
    const effectsGrid = document.getElementById('effects-chips');
    const modifiersGrid = document.getElementById('modifiers-chips');
    const tagsGrid = document.getElementById('tags-chips');

    // Voices
    voicesGrid.innerHTML = AppState.data.voices.map(v => `
        <div class="chip chip-voice" draggable="true" data-value="(${v.name})" data-type="voice">
            <span class="chip-name">${v.name}</span>
            <button class="preview-btn" data-url="${v.preview_url}" title="Preview voice">
                <svg class="icon-speaker" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02z"/>
                </svg>
                <svg class="icon-stop" viewBox="0 0 24 24" fill="currentColor" style="display:none">
                    <rect x="6" y="6" width="12" height="12"/>
                </svg>
            </button>
        </div>
    `).join('') || '<p class="chip-empty">No voices available</p>';

    // Effects
    effectsGrid.innerHTML = AppState.data.effects.map(e => `
        <div class="chip chip-effect" draggable="true" data-value="(${e})" data-type="effect">
            <span class="chip-name">${e}</span>
            <button class="preview-btn" data-url="/effects/${e}.mp3" title="Preview effect">
                <svg class="icon-speaker" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M3 9v6h4l5 5V4L7 9H3zm13.5 3c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02z"/>
                </svg>
                <svg class="icon-stop" viewBox="0 0 24 24" fill="currentColor" style="display:none">
                    <rect x="6" y="6" width="12" height="12"/>
                </svg>
            </button>
        </div>
    `).join('') || '<p class="chip-empty">No effects available</p>';

    // Modifiers
    modifiersGrid.innerHTML = AppState.data.modifiers.map(m => `
        <div class="chip chip-modifier" draggable="true" data-value="(${m})" data-type="modifier">
            <span class="chip-name">${m}</span>
        </div>
    `).join('') || '<p class="chip-empty">No modifiers available</p>';

    // Tags
    tagsGrid.innerHTML = AppState.data.tags.map(t => `
        <div class="chip chip-tag" draggable="true" data-value="[${t}]" data-type="tag">
            <span class="chip-name">${t}</span>
        </div>
    `).join('') || '<p class="chip-empty">No tags available</p>';

    // Reinitialize drag/drop and preview after populating
    initDragDrop();
    initPreviewButtons();
}

function populateAudioLists() {
    const voicesList = document.getElementById('voices-list');
    const effectsList = document.getElementById('effects-list');

    // Voices list
    voicesList.innerHTML = AppState.data.voices.map(v => `
        <li class="audio-item">
            <div class="audio-item-name cyan">${v.name}</div>
            <audio controls>
                <source src="${v.preview_url}" type="audio/mpeg">
            </audio>
        </li>
    `).join('') || '<p class="chip-empty">No voices available</p>';

    // Effects list
    effectsList.innerHTML = AppState.data.effects.map(e => `
        <li class="audio-item">
            <div class="audio-item-name pink">${e}</div>
            <audio controls>
                <source src="/effects/${e}.mp3" type="audio/mpeg">
            </audio>
        </li>
    `).join('') || '<p class="chip-empty">No effects available</p>';
}

// =============================================
// SPA Router
// =============================================

function initRouter() {
    // Handle clicks on SPA links
    document.querySelectorAll('[data-spa-link]').forEach(link => {
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const href = link.getAttribute('href');
            navigateTo(href, true);
        });
    });

    // Handle browser back/forward
    window.addEventListener('popstate', (e) => {
        const path = window.location.pathname;
        navigateTo(path, false);
    });
}

function navigateTo(path, pushState = true) {
    // Map paths to page names
    const pageMap = {
        '/create': 'create',
        '/voices': 'voices',
        '/effects': 'effects',
        '/chart': 'chart',
        '/usage': 'chart'
    };

    const pageName = pageMap[path] || 'create';

    if (AppState.currentPage === pageName) return;

    // Update URL if needed
    if (pushState) {
        history.pushState({ page: pageName }, '', path);
    }

    // Transition pages
    transitionToPage(pageName);

    // Update nav links
    document.querySelectorAll('.nav-link').forEach(link => {
        link.classList.remove('active');
        if (link.dataset.page === pageName) {
            link.classList.add('active');
        }
    });

    // Update page title
    const titles = {
        create: 'TTS Message Creator - Cyan TTS',
        voices: 'Voices - Cyan TTS',
        effects: 'Effects - Cyan TTS',
        chart: 'Usage - Cyan TTS'
    };
    document.title = titles[pageName] || 'Cyan TTS';

    AppState.currentPage = pageName;
}

function transitionToPage(pageName) {
    const pages = document.querySelectorAll('.page');
    const targetPage = document.getElementById(`page-${pageName}`);

    // Fade out current page
    pages.forEach(page => {
        if (page.classList.contains('visible')) {
            page.classList.add('fade-out');
            page.classList.remove('visible');

            setTimeout(() => {
                page.classList.remove('active', 'fade-out');
            }, 300);
        }
    });

    // Fade in new page
    setTimeout(() => {
        targetPage.classList.add('active');

        // Trigger reflow for animation
        targetPage.offsetHeight;

        requestAnimationFrame(() => {
            targetPage.classList.add('visible');
        });
    }, 150);
}

// =============================================
// Create Page Functionality
// =============================================

let dropIndicator = null;

function initCreatePage() {
    const composer = document.getElementById('composer');
    const copyBtn = document.getElementById('copyBtn');
    const clearBtn = document.getElementById('clearBtn');

    // Tab switching
    document.querySelectorAll('.toolbox-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;

            document.querySelectorAll('.toolbox-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            document.querySelectorAll('.chip-panel').forEach(p => p.classList.remove('active'));
            document.getElementById(`panel-${tabName}`).classList.add('active');
        });
    });

    // Composer input handling
    composer.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') e.preventDefault();
    });

    composer.addEventListener('paste', (e) => {
        e.preventDefault();
        const text = (e.clipboardData || window.clipboardData).getData('text/plain');
        const cleanText = text.replace(/[\r\n]+/g, ' ').replace(/\s+/g, ' ');
        document.execCommand('insertText', false, cleanText);
    });

    composer.addEventListener('input', () => {
        updatePlaceholder();
        validateMessage();
    });

    composer.addEventListener('focus', updatePlaceholder);
    composer.addEventListener('blur', updatePlaceholder);

    // Actions
    copyBtn.addEventListener('click', copyToClipboard);
    clearBtn.addEventListener('click', () => {
        composer.textContent = '';
        updatePlaceholder();
        validateMessage();
        composer.focus();
    });

    // Volume slider
    const volumeSlider = document.getElementById('previewVolume');
    const volumeValue = document.getElementById('volumeValue');

    volumeSlider.addEventListener('input', () => {
        const value = volumeSlider.value;
        AppState.previewVolume = value / 100;
        volumeValue.textContent = value + '%';

        // Update currently playing audio if any
        if (AppState.currentAudio) {
            AppState.currentAudio.volume = AppState.previewVolume;
        }
    });

    // Add placeholder styles
    const style = document.createElement('style');
    style.textContent = `
        .message-composer[data-empty="true"]::before {
            content: attr(data-placeholder);
            color: var(--text-muted);
            pointer-events: none;
        }
    `;
    document.head.appendChild(style);
}

function initDragDrop() {
    const composer = document.getElementById('composer');

    document.querySelectorAll('.chip').forEach(chip => {
        chip.addEventListener('dragstart', (e) => {
            e.target.classList.add('dragging');
            e.dataTransfer.setData('text/plain', e.target.dataset.value);
            e.dataTransfer.effectAllowed = 'copy';
        });

        chip.addEventListener('dragend', (e) => {
            e.target.classList.remove('dragging');
            removeDropIndicator();
        });
    });

    composer.addEventListener('dragover', (e) => {
        e.preventDefault();
        e.dataTransfer.dropEffect = 'copy';
        composer.classList.add('drag-over');
        showDropIndicator(e.clientX, e.clientY);
    });

    composer.addEventListener('dragleave', (e) => {
        if (!composer.contains(e.relatedTarget)) {
            composer.classList.remove('drag-over');
            removeDropIndicator();
        }
    });

    composer.addEventListener('drop', (e) => {
        e.preventDefault();
        composer.classList.remove('drag-over');

        const chipValue = e.dataTransfer.getData('text/plain');
        const range = getCaretRangeFromPoint(e.clientX, e.clientY);

        removeDropIndicator();

        if (range) {
            const selection = window.getSelection();
            selection.removeAllRanges();
            selection.addRange(range);

            const text = getComposerText();
            const insertPos = getCaretOffset(composer, range);
            let insertText = chipValue;

            if (insertPos > 0 && text[insertPos - 1] !== ' ') {
                insertText = ' ' + insertText;
            }
            if (insertPos < text.length && text[insertPos] !== ' ') {
                insertText = insertText + ' ';
            }

            document.execCommand('insertText', false, insertText);
        } else {
            const currentText = getComposerText();
            const space = currentText && !currentText.endsWith(' ') ? ' ' : '';
            composer.textContent = currentText + space + chipValue + ' ';
            placeCaretAtEnd(composer);
        }

        updatePlaceholder();
        validateMessage();
    });
}

function showDropIndicator(x, y) {
    const range = getCaretRangeFromPoint(x, y);

    if (range) {
        removeDropIndicator();

        dropIndicator = document.createElement('span');
        dropIndicator.className = 'drop-indicator';
        dropIndicator.id = 'drop-indicator';

        range.insertNode(dropIndicator);
    }
}

function removeDropIndicator() {
    if (dropIndicator && dropIndicator.parentNode) {
        dropIndicator.parentNode.removeChild(dropIndicator);
    }
    dropIndicator = null;

    const existing = document.getElementById('drop-indicator');
    if (existing) existing.parentNode.removeChild(existing);
}

function getCaretRangeFromPoint(x, y) {
    if (document.caretRangeFromPoint) {
        return document.caretRangeFromPoint(x, y);
    } else if (document.caretPositionFromPoint) {
        const pos = document.caretPositionFromPoint(x, y);
        if (pos) {
            const range = document.createRange();
            range.setStart(pos.offsetNode, pos.offset);
            range.collapse(true);
            return range;
        }
    }
    return null;
}

function getCaretOffset(element, range) {
    const preCaretRange = range.cloneRange();
    preCaretRange.selectNodeContents(element);
    preCaretRange.setEnd(range.startContainer, range.startOffset);
    return preCaretRange.toString().length;
}

function placeCaretAtEnd(el) {
    el.focus();
    const range = document.createRange();
    range.selectNodeContents(el);
    range.collapse(false);
    const sel = window.getSelection();
    sel.removeAllRanges();
    sel.addRange(range);
}

function getComposerText() {
    const composer = document.getElementById('composer');
    const clone = composer.cloneNode(true);
    const indicator = clone.querySelector('#drop-indicator');
    if (indicator) indicator.remove();
    return clone.textContent || '';
}

function updatePlaceholder() {
    const composer = document.getElementById('composer');
    const text = getComposerText().trim();
    if (text === '') {
        composer.setAttribute('data-empty', 'true');
    } else {
        composer.removeAttribute('data-empty');
    }
}

function validateMessage() {
    const text = getComposerText().trim();
    const validation = document.getElementById('validation');
    const copyBtn = document.getElementById('copyBtn');

    if (!text) {
        validation.innerHTML = '';
        validation.className = 'validation-status';
        copyBtn.disabled = true;
        return { valid: false, error: null, silent: true };
    }

    const tagRegex = /\(([^)]+)\)/g;
    const matches = [...text.matchAll(tagRegex)];

    const voices = new Set(AppState.data.voices.map(v => v.name.toLowerCase()));
    const effects = new Set(AppState.data.effects.map(e => e.toLowerCase()));
    const modifiers = new Set(AppState.data.modifiers.map(m => m.toLowerCase()));

    let lastTagEnd = 0;
    let pendingVoice = null;
    let voiceHasText = false;

    for (const match of matches) {
        const tagContent = match[1].toLowerCase();
        const tagStart = match.index;
        const tagEnd = tagStart + match[0].length;
        const textBetween = text.substring(lastTagEnd, tagStart).trim();

        if (textBetween !== '' && pendingVoice !== null) {
            voiceHasText = true;
        }

        if (voices.has(tagContent)) {
            if (pendingVoice !== null && !voiceHasText) {
                showValidation({ valid: false, error: `Voice "${pendingVoice}" has no text after it` });
                return { valid: false, error: `Voice "${pendingVoice}" has no text after it` };
            }
            pendingVoice = tagContent;
            voiceHasText = false;
        } else if (!effects.has(tagContent) && !modifiers.has(tagContent)) {
            showValidation({ valid: false, error: `Unknown tag: "${tagContent}"` });
            return { valid: false, error: `Unknown tag: "${tagContent}"` };
        }

        lastTagEnd = tagEnd;
    }

    if (pendingVoice !== null) {
        const textAfter = text.substring(lastTagEnd).trim().replace(/\[[^\]]+\]/g, '').trim();
        if (!voiceHasText && textAfter === '') {
            showValidation({ valid: false, error: `Voice "${pendingVoice}" has no text after it` });
            return { valid: false, error: `Voice "${pendingVoice}" has no text after it` };
        }
    }

    showValidation({ valid: true, error: null });
    return { valid: true, error: null };
}

function showValidation(result) {
    const validation = document.getElementById('validation');
    const copyBtn = document.getElementById('copyBtn');

    if (result.valid) {
        validation.innerHTML = `<span class="validation-icon">✓</span><span class="validation-message">Message is valid</span>`;
        validation.className = 'validation-status validation-valid';
        copyBtn.disabled = false;
    } else {
        validation.innerHTML = `<span class="validation-icon">✕</span><span class="validation-message">${result.error}</span>`;
        validation.className = 'validation-status validation-invalid';
        copyBtn.disabled = true;
    }
}

function initPreviewButtons() {
    document.querySelectorAll('.preview-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            togglePreview(btn, btn.dataset.url);
        });
    });
}

function togglePreview(button, audioUrl) {
    if (AppState.currentButton === button && AppState.currentAudio) {
        stopCurrentAudio();
        return;
    }

    stopCurrentAudio();

    AppState.currentAudio = new Audio(audioUrl);
    AppState.currentAudio.volume = AppState.previewVolume;
    AppState.currentButton = button;

    button.classList.add('playing');
    button.querySelector('.icon-speaker').style.display = 'none';
    button.querySelector('.icon-stop').style.display = 'block';

    AppState.currentAudio.play().catch(() => resetButton(button));

    AppState.currentAudio.addEventListener('ended', () => {
        resetButton(button);
        AppState.currentAudio = null;
        AppState.currentButton = null;
    });
}

function stopCurrentAudio() {
    if (AppState.currentAudio) {
        AppState.currentAudio.pause();
        AppState.currentAudio.currentTime = 0;
        AppState.currentAudio = null;
    }
    if (AppState.currentButton) {
        resetButton(AppState.currentButton);
        AppState.currentButton = null;
    }
}

function resetButton(button) {
    button.classList.remove('playing');
    button.querySelector('.icon-speaker').style.display = 'block';
    button.querySelector('.icon-stop').style.display = 'none';
}

async function copyToClipboard() {
    const text = getComposerText().trim();
    if (!text) return;

    try {
        await navigator.clipboard.writeText(text);
        showToast('Message copied to clipboard!');
    } catch {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.cssText = 'position:fixed;opacity:0';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        showToast('Message copied to clipboard!');
    }
}

function showToast(message) {
    const toast = document.getElementById('toast');
    toast.textContent = message;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 3000);
}

// =============================================
// Chart Page Functionality
// =============================================

function initChartPage() {
    const channelInput = document.getElementById('channel-input');
    const loadBtn = document.getElementById('loadDataBtn');

    channelInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') fetchChartData();
    });

    loadBtn.addEventListener('click', fetchChartData);

    // Load initial eleven data
    fetchElevenData();
}

async function fetchElevenData() {
    try {
        const response = await fetch('/eleven/characters');
        if (!response.ok) return;

        const data = await response.json();
        const timeUntilReset = (data.characters_reset * 1000) - Date.now();
        const days = Math.floor(timeUntilReset / 86400000);
        const hours = Math.floor((timeUntilReset % 86400000) / 3600000);
        const minutes = Math.floor((timeUntilReset % 3600000) / 60000);

        let resetTime = 'soon';
        if (days > 0) resetTime = `in ${days} day${days > 1 ? 's' : ''}`;
        else if (hours > 0) resetTime = `in ${hours} hour${hours > 1 ? 's' : ''}`;
        else if (minutes > 0) resetTime = `in ${minutes} minute${minutes > 1 ? 's' : ''}`;

        document.getElementById('characters-left').textContent = data.characters_left.toLocaleString();
        document.getElementById('characters-reset').textContent = resetTime;
    } catch (err) {
        console.error('Error fetching eleven data:', err);
    }
}

async function fetchChartData() {
    const channel = document.getElementById('channel-input').value.toLowerCase();
    if (!channel) {
        showToast('Please enter a channel name');
        return;
    }

    const period = document.getElementById('period-select').value;
    const endDate = new Date().toISOString().split('T')[0];
    let startDate;

    if (period === 'custom') {
        startDate = document.getElementById('start-date').value;
        const endVal = document.getElementById('end-date').value;
        if (!startDate || !endVal) {
            showToast('Please select a valid date range');
            return;
        }
    } else {
        const days = { week: 7, month: 30, year: 365 }[period];
        startDate = new Date(Date.now() - days * 86400000).toISOString().split('T')[0];
    }

    try {
        const response = await fetch(`/data/${channel}?start=${startDate}&end=${endDate}`);
        if (!response.ok) throw new Error('Failed to fetch data');

        const data = await response.json();

        // Group by date
        const grouped = data.reduce((acc, curr) => {
            const date = curr.date.split('T')[0];
            if (acc[date]) {
                acc[date].chars += curr.num_characters;
                acc[date].cost += curr.estimated_cost;
            } else {
                acc[date] = { chars: curr.num_characters, cost: curr.estimated_cost };
            }
            return acc;
        }, {});

        const sorted = Object.entries(grouped)
            .sort(([a], [b]) => new Date(a) - new Date(b))
            .map(([date, vals]) => ({ date, ...vals }));

        const totalChars = sorted.reduce((s, d) => s + d.chars, 0);
        const totalCost = sorted.reduce((s, d) => s + d.cost, 0);

        await fetchElevenData();

        document.getElementById('total-characters').textContent = totalChars.toLocaleString();
        document.getElementById('total-cost').textContent = '$' + totalCost.toFixed(2);

        document.getElementById('chart-container').style.display = 'block';

        if (AppState.chart) AppState.chart.destroy();

        const ctx = document.getElementById('dataChart').getContext('2d');
        AppState.chart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: sorted.map(d => d.date),
                datasets: [{
                    label: 'Characters',
                    data: sorted.map(d => d.chars),
                    backgroundColor: 'rgba(0, 212, 255, 0.4)',
                    borderColor: 'rgba(0, 212, 255, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { labels: { color: '#fff', font: { size: 14 } } },
                    tooltip: {
                        callbacks: {
                            label: (ctx) => {
                                const d = sorted[ctx.dataIndex];
                                return `Characters: ${d.chars.toLocaleString()} | Cost: $${d.cost.toFixed(2)}`;
                            }
                        }
                    }
                },
                scales: {
                    x: { ticks: { color: 'rgba(255,255,255,0.7)' }, grid: { display: false } },
                    y: { ticks: { color: 'rgba(255,255,255,0.7)' }, grid: { color: 'rgba(255,255,255,0.1)' } }
                }
            }
        });
    } catch (err) {
        console.error('Error:', err);
        showToast('Failed to load data');
    }
}
