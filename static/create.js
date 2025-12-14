/**
 * TTS Message Creator - JavaScript
 * Handles drag-drop, audio preview, validation, and clipboard
 */

// =============================================
// State
// =============================================

let currentAudio = null;
let currentButton = null;
let dropIndicator = null;

// =============================================
// DOM Elements
// =============================================

const composer = document.getElementById('composer');
const validation = document.getElementById('validation');
const copyBtn = document.getElementById('copyBtn');
const clearBtn = document.getElementById('clearBtn');
const toast = document.getElementById('toast');
const tabs = document.querySelectorAll('.toolbox-tab');
const panels = document.querySelectorAll('.chip-panel');
const chips = document.querySelectorAll('.chip');
const previewBtns = document.querySelectorAll('.preview-btn');

// =============================================
// Initialization
// =============================================

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initDragDrop();
    initComposer();
    initPreviewButtons();
    initActions();
    updatePlaceholder();
});

// =============================================
// Tab Switching
// =============================================

function initTabs() {
    tabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;

            // Update active tab
            tabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');

            // Show corresponding panel
            panels.forEach(p => p.classList.remove('active'));
            document.getElementById(`panel-${tabName}`).classList.add('active');
        });
    });
}

// =============================================
// Drag and Drop
// =============================================

function initDragDrop() {
    chips.forEach(chip => {
        chip.addEventListener('dragstart', handleDragStart);
        chip.addEventListener('dragend', handleDragEnd);
    });

    composer.addEventListener('dragover', handleDragOver);
    composer.addEventListener('dragleave', handleDragLeave);
    composer.addEventListener('drop', handleDrop);
}

function handleDragStart(e) {
    e.target.classList.add('dragging');
    e.dataTransfer.setData('text/plain', e.target.dataset.value);
    e.dataTransfer.effectAllowed = 'copy';
}

function handleDragEnd(e) {
    e.target.classList.remove('dragging');
    removeDropIndicator();
}

function handleDragOver(e) {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    composer.classList.add('drag-over');

    // Show drop indicator at cursor position
    showDropIndicator(e.clientX, e.clientY);
}

function handleDragLeave(e) {
    // Only remove if actually leaving the composer
    if (!composer.contains(e.relatedTarget)) {
        composer.classList.remove('drag-over');
        removeDropIndicator();
    }
}

function handleDrop(e) {
    e.preventDefault();
    composer.classList.remove('drag-over');

    const chipValue = e.dataTransfer.getData('text/plain');

    // Get the position where we should insert
    const range = getCaretRangeFromPoint(e.clientX, e.clientY);

    if (range) {
        // Remove drop indicator first
        removeDropIndicator();

        // Insert the chip value at the caret position
        const selection = window.getSelection();
        selection.removeAllRanges();
        selection.addRange(range);

        // Add space before/after if needed
        const text = getComposerText();
        const insertPos = getCaretOffset(composer, range);
        let insertText = chipValue;

        // Add space before if not at start and previous char isn't space
        if (insertPos > 0 && text[insertPos - 1] !== ' ') {
            insertText = ' ' + insertText;
        }
        // Add space after
        if (insertPos < text.length && text[insertPos] !== ' ') {
            insertText = insertText + ' ';
        }

        document.execCommand('insertText', false, insertText);
    } else {
        // Fallback: append to end
        removeDropIndicator();
        const currentText = getComposerText();
        const space = currentText && !currentText.endsWith(' ') ? ' ' : '';
        composer.textContent = currentText + space + chipValue + ' ';
        placeCaretAtEnd(composer);
    }

    updatePlaceholder();
    validateMessage();
}

function showDropIndicator(x, y) {
    const range = getCaretRangeFromPoint(x, y);

    if (range) {
        // Remove existing indicator
        removeDropIndicator();

        // Create new indicator
        dropIndicator = document.createElement('span');
        dropIndicator.className = 'drop-indicator';
        dropIndicator.id = 'drop-indicator';

        // Insert at the caret position
        range.insertNode(dropIndicator);
    }
}

function removeDropIndicator() {
    if (dropIndicator && dropIndicator.parentNode) {
        dropIndicator.parentNode.removeChild(dropIndicator);
    }
    dropIndicator = null;

    // Also try to find by ID in case reference was lost
    const existing = document.getElementById('drop-indicator');
    if (existing) {
        existing.parentNode.removeChild(existing);
    }
}

function getCaretRangeFromPoint(x, y) {
    let range;

    if (document.caretRangeFromPoint) {
        // Chrome/Safari
        range = document.caretRangeFromPoint(x, y);
    } else if (document.caretPositionFromPoint) {
        // Firefox
        const pos = document.caretPositionFromPoint(x, y);
        if (pos) {
            range = document.createRange();
            range.setStart(pos.offsetNode, pos.offset);
            range.collapse(true);
        }
    }

    return range;
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

// =============================================
// Composer Input Handling
// =============================================

function initComposer() {
    // Prevent newlines
    composer.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
        }
    });

    // Handle paste - strip newlines
    composer.addEventListener('paste', (e) => {
        e.preventDefault();
        const text = (e.clipboardData || window.clipboardData).getData('text/plain');
        const cleanText = text.replace(/[\r\n]+/g, ' ').replace(/\s+/g, ' ');
        document.execCommand('insertText', false, cleanText);
    });

    // Update validation on input
    composer.addEventListener('input', () => {
        updatePlaceholder();
        validateMessage();
    });

    // Handle focus for placeholder
    composer.addEventListener('focus', updatePlaceholder);
    composer.addEventListener('blur', updatePlaceholder);
}

function getComposerText() {
    // Get text without drop indicator
    const clone = composer.cloneNode(true);
    const indicator = clone.querySelector('#drop-indicator');
    if (indicator) {
        indicator.remove();
    }
    return clone.textContent || '';
}

function updatePlaceholder() {
    const text = getComposerText().trim();
    if (text === '') {
        composer.setAttribute('data-empty', 'true');
    } else {
        composer.removeAttribute('data-empty');
    }
}

// =============================================
// Validation
// =============================================

function validateMessage() {
    const text = getComposerText().trim();

    // Empty message - silent invalid (no error shown)
    if (!text) {
        validation.innerHTML = '';
        validation.className = 'validation-status';
        copyBtn.disabled = true;
        return { valid: false, error: null, silent: true };
    }

    // Parse tags
    const tagRegex = /\(([^)]+)\)/g;
    const matches = [...text.matchAll(tagRegex)];

    // Get configured voices, effects, modifiers from page data
    const voiceChips = document.querySelectorAll('.chip-voice');
    const effectChips = document.querySelectorAll('.chip-effect');
    const modifierChips = document.querySelectorAll('.chip-modifier');

    const voices = new Set();
    const effects = new Set();
    const modifiers = new Set();

    voiceChips.forEach(c => voices.add(c.dataset.value.replace(/[()]/g, '').toLowerCase()));
    effectChips.forEach(c => effects.add(c.dataset.value.replace(/[()]/g, '').toLowerCase()));
    modifierChips.forEach(c => modifiers.add(c.dataset.value.replace(/[()]/g, '').toLowerCase()));

    // Track state
    let lastTagEnd = 0;
    let pendingVoice = null;      // Voice waiting for text
    let voiceHasText = false;     // Whether pending voice has received any text

    for (let i = 0; i < matches.length; i++) {
        const match = matches[i];
        const tagContent = match[1].toLowerCase();
        const tagStart = match.index;
        const tagEnd = tagStart + match[0].length;

        // Check text between last tag and this one
        const textBetween = text.substring(lastTagEnd, tagStart).trim();

        // If there's text between tags, the pending voice now has text
        if (textBetween !== '' && pendingVoice !== null) {
            voiceHasText = true;
        }

        // Determine tag type
        if (voices.has(tagContent)) {
            // Voice tag - check if previous voice had text
            if (pendingVoice !== null && !voiceHasText) {
                const result = {
                    valid: false,
                    error: `Voice "${pendingVoice}" has no text after it`,
                    silent: false
                };
                showValidation(result);
                return result;
            }
            // Start tracking new voice
            pendingVoice = tagContent;
            voiceHasText = false;
        } else if (effects.has(tagContent)) {
            // Effect - always valid, doesn't affect voice tracking
            // (voice can have text before or after effects)
        } else if (modifiers.has(tagContent)) {
            // Modifier - doesn't affect voice tracking
        } else {
            // Unknown tag
            const result = {
                valid: false,
                error: `Unknown tag: "${tagContent}"`,
                silent: false
            };
            showValidation(result);
            return result;
        }

        lastTagEnd = tagEnd;
    }

    // Check if last voice has text after it
    if (pendingVoice !== null) {
        const textAfter = text.substring(lastTagEnd).trim();
        // Remove any ElevenLabs tags [...] from the text
        const cleanTextAfter = textAfter.replace(/\[[^\]]+\]/g, '').trim();

        // Voice is valid if it had text before OR has text after
        if (!voiceHasText && cleanTextAfter === '') {
            const result = {
                valid: false,
                error: `Voice "${pendingVoice}" has no text after it`,
                silent: false
            };
            showValidation(result);
            return result;
        }
    }

    // Check if there's any text without a voice
    if (matches.length === 0) {
        // Just plain text - valid (uses default voice)
        const result = { valid: true, error: null, silent: false };
        showValidation(result);
        return result;
    }

    // All checks passed
    const result = { valid: true, error: null, silent: false };
    showValidation(result);
    return result;
}

function showValidation(result) {
    if (result.silent) {
        validation.innerHTML = '';
        validation.className = 'validation-status';
        copyBtn.disabled = true;
        return;
    }

    if (result.valid) {
        validation.innerHTML = `
            <span class="validation-icon">✓</span>
            <span class="validation-message">Message is valid</span>
        `;
        validation.className = 'validation-status validation-valid';
        copyBtn.disabled = false;
    } else {
        validation.innerHTML = `
            <span class="validation-icon">✕</span>
            <span class="validation-message">${result.error}</span>
        `;
        validation.className = 'validation-status validation-invalid';
        copyBtn.disabled = true;
    }
}

// =============================================
// Audio Preview
// =============================================

function initPreviewButtons() {
    previewBtns.forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation(); // Don't trigger drag
            const url = btn.dataset.url;
            togglePreview(btn, url);
        });
    });
}

function togglePreview(button, audioUrl) {
    // If clicking the same button that's playing, stop it
    if (currentButton === button && currentAudio) {
        stopCurrentAudio();
        return;
    }

    // Stop any currently playing audio
    stopCurrentAudio();

    // Play new audio
    currentAudio = new Audio(audioUrl);
    currentButton = button;

    // Update button state to playing
    button.classList.add('playing');
    button.querySelector('.icon-speaker').style.display = 'none';
    button.querySelector('.icon-stop').style.display = 'block';

    currentAudio.play().catch(err => {
        console.error('Error playing audio:', err);
        resetButton(button);
    });

    currentAudio.addEventListener('ended', () => {
        resetButton(button);
        currentAudio = null;
        currentButton = null;
    });

    currentAudio.addEventListener('error', () => {
        resetButton(button);
        currentAudio = null;
        currentButton = null;
    });
}

function stopCurrentAudio() {
    if (currentAudio) {
        currentAudio.pause();
        currentAudio.currentTime = 0;
        currentAudio = null;
    }
    if (currentButton) {
        resetButton(currentButton);
        currentButton = null;
    }
}

function resetButton(button) {
    button.classList.remove('playing');
    button.querySelector('.icon-speaker').style.display = 'block';
    button.querySelector('.icon-stop').style.display = 'none';
}

// =============================================
// Actions (Copy & Clear)
// =============================================

function initActions() {
    copyBtn.addEventListener('click', copyToClipboard);
    clearBtn.addEventListener('click', clearComposer);
}

async function copyToClipboard() {
    const text = getComposerText().trim();

    if (!text) return;

    try {
        await navigator.clipboard.writeText(text);
        showToast('Message copied to clipboard!');
    } catch (err) {
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = text;
        textArea.style.position = 'fixed';
        textArea.style.opacity = '0';
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand('copy');
        document.body.removeChild(textArea);
        showToast('Message copied to clipboard!');
    }
}

function clearComposer() {
    composer.textContent = '';
    updatePlaceholder();
    validateMessage();
    composer.focus();
}

function showToast(message) {
    toast.textContent = message;
    toast.classList.add('show');

    setTimeout(() => {
        toast.classList.remove('show');
    }, 3000);
}

// =============================================
// Placeholder CSS (since contenteditable doesn't support placeholder)
// =============================================

// Add placeholder styles dynamically
const style = document.createElement('style');
style.textContent = `
    .message-composer[data-empty="true"]::before {
        content: attr(data-placeholder);
        color: var(--text-muted);
        pointer-events: none;
    }
`;
document.head.appendChild(style);
