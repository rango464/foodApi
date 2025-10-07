// Конфигурация WebSocket
const wsaddr = "ws://localhost:1323/ws";

// Основной класс биржевого терминала
class TradingTerminal {
    constructor() {
        // WebSocket соединение
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        
        // FPS счетчик
        this.messageCount = 0;
        this.lastSecond = Date.now();
        this.fps = 0;
        
        // Данные валютных пар
        this.availablePairs = ['BTCUSDT', 'ETHUSDT', 'LTCBTC', 'ADAUSDT', 'BNBUSDT', 'XRPUSDT', 'SOLUSDT', 'DOTUSDT', 'LINKUSDT', 'MATICUSDT', 'AVAXUSDT', 'UNIUSDT'];
        this.selectedPairs = new Set(['BTCUSDT']);
        this.currentPair = 'BTCUSDT';
        this.priceData = {};
        this.filteredPairs = [...this.availablePairs];
        
        // Данные графика
        this.candleData = {};
        this.volumeData = {};
        this.trades = [];
        this.orderBook = {};
        
        // График
        this.candleChart = null;
        this.volumeChart = null;
        this.chartTimeframe = '1m';
        
        // Элементы DOM
        this.elements = {};
        
        // Инициализация
        this.initializeElements();
        this.setupEventListeners();
        this.setupCharts();
        this.generateMockData();
        this.connectWebSocket();
        this.startUpdateLoop();
    }

    // Инициализация элементов DOM
    initializeElements() {
        this.elements = {
            // Статус соединения
            statusIndicator: document.getElementById('statusIndicator'),
            statusText: document.getElementById('statusText'),
            fpsCounter: document.getElementById('fpsCounter'),
            
            // Селектор валютных пар
            pairDropdown: document.getElementById('pairDropdown'),
            dropdownToggle: document.getElementById('dropdownToggle'),
            dropdownMenu: document.getElementById('dropdownMenu'),
            selectedPairsText: document.getElementById('selectedPairsText'),
            selectedCount: document.getElementById('selectedCount'),
            filterInput: document.getElementById('filterInput'),
            selectAllBtn: document.getElementById('selectAllBtn'),
            clearAllBtn: document.getElementById('clearAllBtn'),
            pairCheckboxes: document.getElementById('pairCheckboxes'),
            
            // Вкладки пар
            pairTabs: document.getElementById('pairTabs'),
            
            // График свечей
            candlestickChart: document.getElementById('candlestickChart'),
            volumeChart: document.getElementById('volumeChart'),
            chartCrosshair: document.getElementById('chartCrosshair'),
            priceLabel: document.getElementById('priceLabel'),
            timeLabel: document.getElementById('timeLabel'),
            chartTooltip: document.getElementById('chartTooltip'),
            
            // Панель масштабирования
            timeframeBtns: document.querySelectorAll('.timeframe-btn'),
            
            // История сделок
            totalProfit: document.getElementById('totalProfit'),
            profitAmount: document.getElementById('profitAmount'),
            tradesList: document.getElementById('tradesList'),
            
            // Биржевой стакан
            orderBook: document.getElementById('orderBook'),
            asks: document.getElementById('asks'),
            bids: document.getElementById('bids'),
            spread: document.getElementById('spread'),
            
            // Модальное окно сделки
            tradeModal: document.getElementById('tradeModal'),
            closeModal: document.getElementById('closeModal'),
            tradeForm: document.getElementById('tradeForm'),
            tradeType: document.getElementById('tradeType'),
            tradePair: document.getElementById('tradePair'),
            tradePrice: document.getElementById('tradePrice'),
            tradeAmount: document.getElementById('tradeAmount'),
            
            // Изменение размера
            rightColumn: document.getElementById('rightColumn'),
            resizeHandle: document.getElementById('resizeHandle')
        };
    }

    // Настройка обработчиков событий
    setupEventListeners() {
        // Панель масштабирования
        this.elements.timeframeBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.changeTimeframe(e.target.dataset.timeframe);
            });
        });

        // График - события мыши
        this.elements.candlestickChart.addEventListener('mousemove', (e) => {
            this.handleChartMouseMove(e);
        });

        this.elements.candlestickChart.addEventListener('mouseleave', () => {
            this.hideChartCrosshair();
        });

        this.elements.candlestickChart.addEventListener('dblclick', (e) => {
            this.openTradeModal(e);
        });

        // Модальное окно
        this.elements.closeModal.addEventListener('click', () => {
            this.closeTradeModal();
        });

        this.elements.tradeModal.addEventListener('click', (e) => {
            if (e.target === this.elements.tradeModal) {
                this.closeTradeModal();
            }
        });

        this.elements.tradeForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.executeTrade();
        });

        // Выпадающий список валютных пар
        this.setupPairDropdown();

        // Изменение размера правой колонки
        this.setupResizeHandle();

        // Обновление размера canvas при изменении размера окна
        window.addEventListener('resize', () => {
            this.resizeCharts();
        });

        // Закрытие выпадающего списка при клике вне его
        document.addEventListener('click', (e) => {
            if (!this.elements.pairDropdown.contains(e.target)) {
                this.closeDropdown();
            }
        });
    }

    // Настройка выпадающего списка валютных пар
    setupPairDropdown() {
        // Обработчик клика по переключателю
        this.elements.dropdownToggle.addEventListener('click', (e) => {
            e.stopPropagation();
            this.toggleDropdown();
        });

        // Обработчик поиска
        this.elements.filterInput.addEventListener('input', (e) => {
            this.filterPairs(e.target.value);
        });

        // Обработчики кнопок "Выбрать все" и "Очистить"
        this.elements.selectAllBtn.addEventListener('click', () => {
            this.selectAllPairs();
        });

        this.elements.clearAllBtn.addEventListener('click', () => {
            this.clearAllPairs();
        });

        // Создание чекбоксов
        this.createPairCheckboxes();
        this.updateDropdownText();
    }

    // Переключение состояния выпадающего списка
    toggleDropdown() {
        const isOpen = this.elements.dropdownMenu.classList.contains('show');
        if (isOpen) {
            this.closeDropdown();
        } else {
            this.openDropdown();
        }
    }

    // Открытие выпадающего списка
    openDropdown() {
        this.elements.dropdownMenu.classList.add('show');
        this.elements.dropdownToggle.classList.add('active');
        this.elements.filterInput.focus();
    }

    // Закрытие выпадающего списка
    closeDropdown() {
        this.elements.dropdownMenu.classList.remove('show');
        this.elements.dropdownToggle.classList.remove('active');
        this.elements.filterInput.value = '';
        this.filterPairs(''); // Сбросить фильтр
    }

    // Фильтрация валютных пар
    filterPairs(searchTerm) {
        const term = searchTerm.toLowerCase();
        this.filteredPairs = this.availablePairs.filter(pair => 
            pair.toLowerCase().includes(term)
        );
        this.updatePairCheckboxes();
    }

    // Выбрать все видимые пары
    selectAllPairs() {
        this.filteredPairs.forEach(pair => {
            this.selectedPairs.add(pair);
        });
        this.updatePairCheckboxes();
        this.updateDropdownText();
        this.updateSelectedCount();
        this.updatePairTabs();
    }

    // Очистить все выбранные пары
    clearAllPairs() {
        // Оставляем хотя бы одну пару выбранной
        if (this.selectedPairs.size > 1) {
            this.selectedPairs.clear();
            this.selectedPairs.add(this.currentPair);
        }
        this.updatePairCheckboxes();
        this.updateDropdownText();
        this.updateSelectedCount();
        this.updatePairTabs();
    }

    // Настройка изменения размера
    setupResizeHandle() {
        let isResizing = false;
        let startX = 0;
        let startWidth = 0;

        this.elements.resizeHandle.addEventListener('mousedown', (e) => {
            isResizing = true;
            startX = e.clientX;
            startWidth = this.elements.rightColumn.offsetWidth;
            document.addEventListener('mousemove', handleResize);
            document.addEventListener('mouseup', stopResize);
            e.preventDefault();
        });

        const handleResize = (e) => {
            if (!isResizing) return;
            const deltaX = startX - e.clientX;
            const newWidth = Math.max(200, Math.min(600, startWidth + deltaX));
            this.elements.rightColumn.style.width = newWidth + 'px';
            this.resizeCharts();
        };

        const stopResize = () => {
            isResizing = false;
            document.removeEventListener('mousemove', handleResize);
            document.removeEventListener('mouseup', stopResize);
        };
    }

    // WebSocket подключение
    connectWebSocket() {
        try {
            this.ws = new WebSocket(wsaddr);
            
            this.ws.onopen = () => {
                console.log('WebSocket подключен к', wsaddr);
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.updateConnectionStatus();
            };

            this.ws.onmessage = (event) => {
                this.handleWebSocketMessage(event.data);
            };

            this.ws.onclose = () => {
                console.log('WebSocket отключен');
                this.isConnected = false;
                this.updateConnectionStatus();
                this.attemptReconnect();
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket ошибка:', error);
                this.isConnected = false;
                this.updateConnectionStatus();
            };

        } catch (error) {
            console.error('Ошибка создания WebSocket:', error);
            this.isConnected = false;
            this.updateConnectionStatus();
        }
    }

    // Попытка переподключения
    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            console.log(`Попытка переподключения ${this.reconnectAttempts}/${this.maxReconnectAttempts}`);
            setTimeout(() => {
                this.connectWebSocket();
            }, 2000 * this.reconnectAttempts);
        }
    }

    // Обработка сообщений WebSocket
    handleWebSocketMessage(data) {
        this.messageCount++;
        this.updateFPS();

        try {
            const parsedData = JSON.parse(data);
            this.updatePriceData(parsedData);
        } catch (error) {
            console.log('Получено сообщение:', data);
        }
    }

    // Обновление данных цен
    updatePriceData(data) {
        Object.keys(data).forEach(pair => {
            if (typeof data[pair] === 'number') {
                this.priceData[pair] = data[pair];
                this.updateCandleData(pair, data[pair]);
                this.updateOrderBook(pair, data[pair]);
            }
        });
        
        this.updateUI();
    }

    // Обновление данных свечей
    updateCandleData(pair, price) {
        if (!this.candleData[pair]) {
            this.candleData[pair] = [];
        }

        const now = Date.now();
        const timeframeMs = this.getTimeframeMs(this.chartTimeframe);
        const candleTime = Math.floor(now / timeframeMs) * timeframeMs;

        let lastCandle = this.candleData[pair][this.candleData[pair].length - 1];

        if (!lastCandle || lastCandle.timestamp !== candleTime) {
            // Создаем новую свечу
            const newCandle = {
                timestamp: candleTime,
                open: price,
                high: price,
                low: price,
                close: price,
                volume: Math.random() * 100 + 10
            };
            this.candleData[pair].push(newCandle);
            
            // Ограничиваем количество свечей
            if (this.candleData[pair].length > 200) {
                this.candleData[pair].shift();
            }
        } else {
            // Обновляем текущую свечу
            lastCandle.close = price;
            lastCandle.high = Math.max(lastCandle.high, price);
            lastCandle.low = Math.min(lastCandle.low, price);
            lastCandle.volume += Math.random() * 5;
        }
    }

    // Получение интервала времени в миллисекундах
    getTimeframeMs(timeframe) {
        const timeframes = {
            '1m': 60 * 1000,
            '5m': 5 * 60 * 1000,
            '10m': 10 * 60 * 1000,
            '30m': 30 * 60 * 1000,
            '1h': 60 * 60 * 1000,
            '1d': 24 * 60 * 60 * 1000,
            '1w': 7 * 24 * 60 * 60 * 1000
        };
        return timeframes[timeframe] || timeframes['1m'];
    }

    // Обновление биржевого стакана
    updateOrderBook(pair, price) {
        if (!this.orderBook[pair]) {
            this.orderBook[pair] = { asks: [], bids: [] };
        }

        const book = this.orderBook[pair];
        book.asks = [];
        book.bids = [];

        // Генерируем аски (продажи)
        for (let i = 1; i <= 15; i++) {
            const askPrice = price + (i * (Math.random() * 10 + 1));
            const volume = Math.random() * 5 + 0.1;
            book.asks.push({ price: askPrice, volume: volume });
        }

        // Генерируем биды (покупки)
        for (let i = 1; i <= 15; i++) {
            const bidPrice = price - (i * (Math.random() * 10 + 1));
            const volume = Math.random() * 5 + 0.1;
            book.bids.push({ price: bidPrice, volume: volume });
        }

        book.asks.sort((a, b) => a.price - b.price);
        book.bids.sort((a, b) => b.price - a.price);
    }

    // Обновление FPS счетчика
    updateFPS() {
        const now = Date.now();
        if (now - this.lastSecond >= 1000) {
            this.fps = this.messageCount;
            this.messageCount = 0;
            this.lastSecond = now;
            this.elements.fpsCounter.textContent = `${this.fps} FPS`;
        }
    }

    // Обновление статуса соединения
    updateConnectionStatus() {
        if (this.isConnected) {
            this.elements.statusIndicator.classList.add('connected');
            this.elements.statusText.textContent = 'Подключено';
        } else {
            this.elements.statusIndicator.classList.remove('connected');
            this.elements.statusText.textContent = 'Отключено';
            this.elements.fpsCounter.textContent = '0 FPS';
        }
    }

    // Создание чекбоксов для валютных пар
    createPairCheckboxes() {
        this.elements.pairCheckboxes.innerHTML = '';
        
        this.availablePairs.forEach(pair => {
            const checkbox = document.createElement('div');
            checkbox.className = 'pair-checkbox';
            checkbox.dataset.pair = pair;
            
            const isSelected = this.selectedPairs.has(pair);
            if (isSelected) {
                checkbox.classList.add('selected');
            }
            
            checkbox.innerHTML = `
                <input type="checkbox" id="pair-${pair}" ${isSelected ? 'checked' : ''}>
                <label for="pair-${pair}">${this.formatPairName(pair)}</label>
            `;
            
            const input = checkbox.querySelector('input');
            const label = checkbox.querySelector('label');
            
            // Обработчик изменения чекбокса
            const handleChange = (checked) => {
                if (checked) {
                    this.selectedPairs.add(pair);
                    checkbox.classList.add('selected');
                } else {
                    // Не позволяем убрать все пары
                    if (this.selectedPairs.size > 1) {
                        this.selectedPairs.delete(pair);
                        checkbox.classList.remove('selected');
                    } else {
                        input.checked = true;
                        return;
                    }
                }
                this.updateDropdownText();
                this.updateSelectedCount();
                this.updatePairTabs();
            };
            
            input.addEventListener('change', (e) => {
                handleChange(e.target.checked);
            });
            
            // Клик по всей строке
            checkbox.addEventListener('click', (e) => {
                if (e.target !== input) {
                    input.checked = !input.checked;
                    handleChange(input.checked);
                }
            });
            
            this.elements.pairCheckboxes.appendChild(checkbox);
        });
        
        this.updateSelectedCount();
    }

    // Обновление отображения чекбоксов (для фильтрации)
    updatePairCheckboxes() {
        const checkboxes = this.elements.pairCheckboxes.querySelectorAll('.pair-checkbox');
        
        checkboxes.forEach(checkbox => {
            const pair = checkbox.dataset.pair;
            const isVisible = this.filteredPairs.includes(pair);
            const isSelected = this.selectedPairs.has(pair);
            
            checkbox.style.display = isVisible ? 'flex' : 'none';
            
            if (isSelected) {
                checkbox.classList.add('selected');
            } else {
                checkbox.classList.remove('selected');
            }
            
            const input = checkbox.querySelector('input');
            input.checked = isSelected;
        });
    }

    // Форматирование названия валютной пары
    formatPairName(pair) {
        return pair.replace('USDT', '/USDT').replace('BTC', '/BTC');
    }

    // Обновление текста в dropdown toggle
    updateDropdownText() {
        const selectedArray = Array.from(this.selectedPairs);
        let text = '';
        
        if (selectedArray.length === 0) {
            text = 'Выберите пары';
        } else if (selectedArray.length === 1) {
            text = this.formatPairName(selectedArray[0]);
        } else if (selectedArray.length <= 3) {
            text = selectedArray.map(pair => this.formatPairName(pair)).join(', ');
        } else {
            text = `${this.formatPairName(selectedArray[0])}, ${this.formatPairName(selectedArray[1])} +${selectedArray.length - 2}`;
        }
        
        this.elements.selectedPairsText.textContent = text;
    }

    // Обновление счетчика выбранных пар
    updateSelectedCount() {
        const count = this.selectedPairs.size;
        this.elements.selectedCount.textContent = count > 1 ? count.toString() : '';
        this.elements.selectedCount.style.display = count > 1 ? 'inline-block' : 'none';
    }

    // Создание вкладок для валютных пар
    updatePairTabs() {
        this.elements.pairTabs.innerHTML = '';
        
        Array.from(this.selectedPairs).forEach(pair => {
            const tab = document.createElement('button');
            tab.className = `pair-tab ${pair === this.currentPair ? 'active' : ''}`;
            tab.textContent = this.formatPairName(pair);
            tab.addEventListener('click', () => {
                this.switchToPair(pair);
            });
            this.elements.pairTabs.appendChild(tab);
        });
    }

    // Переключение на валютную пару
    switchToPair(pair) {
        this.currentPair = pair;
        this.updatePairTabs();
        this.updateCharts();
        this.updateOrderBookUI();
    }

    // Изменение таймфрейма
    changeTimeframe(timeframe) {
        this.chartTimeframe = timeframe;
        
        this.elements.timeframeBtns.forEach(btn => {
            btn.classList.remove('active');
            if (btn.dataset.timeframe === timeframe) {
                btn.classList.add('active');
            }
        });
        
        this.regenerateCandleData();
        this.updateCharts();
    }

    // Настройка графиков
    setupCharts() {
        this.resizeCharts();
        this.updatePairTabs();
    }

    // Изменение размера графиков
    resizeCharts() {
        const candleContainer = this.elements.candlestickChart.parentElement;
        const volumeContainer = this.elements.volumeChart.parentElement;
        
        this.elements.candlestickChart.width = candleContainer.clientWidth;
        this.elements.candlestickChart.height = candleContainer.clientHeight;
        
        this.elements.volumeChart.width = volumeContainer.clientWidth;
        this.elements.volumeChart.height = volumeContainer.clientHeight;
        
        this.updateCharts();
    }

    // Обработка движения мыши по графику
    handleChartMouseMove(e) {
        const rect = this.elements.candlestickChart.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        this.showChartCrosshair(x, y);
        this.showChartTooltip(x, y);
    }

    // Показ перекрестия на графике
    showChartCrosshair(x, y) {
        const crosshair = this.elements.chartCrosshair;
        crosshair.innerHTML = `
            <div class="crosshair-line crosshair-vertical" style="left: ${x}px; display: block;"></div>
            <div class="crosshair-line crosshair-horizontal" style="top: ${y}px; display: block;"></div>
        `;
        
        // Показываем метки на осях
        this.elements.priceLabel.style.display = 'block';
        this.elements.priceLabel.style.top = y + 'px';
        
        this.elements.timeLabel.style.display = 'block';
        this.elements.timeLabel.style.left = x + 'px';
        
        // Вычисляем цену и время
        const candleData = this.candleData[this.currentPair] || [];
        if (candleData.length > 0) {
            const chart = this.elements.candlestickChart;
            const priceRange = this.getPriceRange(candleData);
            const price = priceRange.max - ((y / chart.height) * (priceRange.max - priceRange.min));
            
            const timeIndex = Math.floor((x / chart.width) * candleData.length);
            const candle = candleData[Math.max(0, Math.min(timeIndex, candleData.length - 1))];
            
            this.elements.priceLabel.textContent = price.toFixed(2);
            this.elements.timeLabel.textContent = new Date(candle.timestamp).toLocaleTimeString();
        }
    }

    // Показ тултипа графика
    showChartTooltip(x, y) {
        const candleData = this.candleData[this.currentPair] || [];
        if (candleData.length === 0) return;
        
        const chart = this.elements.candlestickChart;
        const timeIndex = Math.floor((x / chart.width) * candleData.length);
        const candle = candleData[Math.max(0, Math.min(timeIndex, candleData.length - 1))];
        
        this.elements.chartTooltip.style.display = 'block';
        this.elements.chartTooltip.style.left = (x + 10) + 'px';
        this.elements.chartTooltip.style.top = (y - 10) + 'px';
        this.elements.chartTooltip.innerHTML = `
            <div><strong>${this.formatPairName(this.currentPair)}</strong></div>
            <div>Время: ${new Date(candle.timestamp).toLocaleString()}</div>
            <div>Открытие: ${candle.open.toFixed(2)}</div>
            <div>Максимум: ${candle.high.toFixed(2)}</div>
            <div>Минимум: ${candle.low.toFixed(2)}</div>
            <div>Закрытие: ${candle.close.toFixed(2)}</div>
            <div>Объем: ${candle.volume.toFixed(2)}</div>
        `;
    }

    // Скрытие перекрестия графика
    hideChartCrosshair() {
        this.elements.chartCrosshair.innerHTML = '';
        this.elements.priceLabel.style.display = 'none';
        this.elements.timeLabel.style.display = 'none';
        this.elements.chartTooltip.style.display = 'none';
    }

    // Получение диапазона цен
    getPriceRange(candleData) {
        if (candleData.length === 0) return { min: 0, max: 100 };
        
        let min = Infinity;
        let max = -Infinity;
        
        candleData.forEach(candle => {
            min = Math.min(min, candle.low);
            max = Math.max(max, candle.high);
        });
        
        const padding = (max - min) * 0.1;
        return { min: min - padding, max: max + padding };
    }

    // Отрисовка графиков
    updateCharts() {
        this.drawCandlestickChart();
        this.drawVolumeChart();
    }

    // Отрисовка графика свечей
    drawCandlestickChart() {
        const canvas = this.elements.candlestickChart;
        const ctx = canvas.getContext('2d');
        const candleData = this.candleData[this.currentPair] || [];
        
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        if (candleData.length === 0) return;
        
        const padding = 20;
        const chartWidth = canvas.width - padding * 2;
        const chartHeight = canvas.height - padding * 2;
        const priceRange = this.getPriceRange(candleData);
        
        // Рисуем фон
        ctx.fillStyle = '#0d1421';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Рисуем сетку
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
        ctx.lineWidth = 1;
        
        for (let i = 0; i <= 10; i++) {
            const y = padding + (i / 10) * chartHeight;
            ctx.beginPath();
            ctx.moveTo(padding, y);
            ctx.lineTo(canvas.width - padding, y);
            ctx.stroke();
        }
        
        for (let i = 0; i <= 10; i++) {
            const x = padding + (i / 10) * chartWidth;
            ctx.beginPath();
            ctx.moveTo(x, padding);
            ctx.lineTo(x, canvas.height - padding);
            ctx.stroke();
        }
        
        // Рисуем свечи
        const candleWidth = Math.max(1, chartWidth / candleData.length * 0.8);
        const candleSpacing = chartWidth / candleData.length;
        
        candleData.forEach((candle, index) => {
            const x = padding + index * candleSpacing + candleSpacing / 2;
            
            const openY = padding + chartHeight - ((candle.open - priceRange.min) / (priceRange.max - priceRange.min)) * chartHeight;
            const closeY = padding + chartHeight - ((candle.close - priceRange.min) / (priceRange.max - priceRange.min)) * chartHeight;
            const highY = padding + chartHeight - ((candle.high - priceRange.min) / (priceRange.max - priceRange.min)) * chartHeight;
            const lowY = padding + chartHeight - ((candle.low - priceRange.min) / (priceRange.max - priceRange.min)) * chartHeight;
            
            const isGreen = candle.close >= candle.open;
            const color = isGreen ? '#38a169' : '#e53e3e';
            
            // Рисуем тень
            ctx.strokeStyle = color;
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.moveTo(x, highY);
            ctx.lineTo(x, lowY);
            ctx.stroke();
            
            // Рисуем тело свечи
            ctx.fillStyle = color;
            const bodyTop = Math.min(openY, closeY);
            const bodyHeight = Math.abs(closeY - openY) || 1;
            ctx.fillRect(x - candleWidth / 2, bodyTop, candleWidth, bodyHeight);
        });
        
        // Рисуем индикаторы сделок
        this.drawTradeIndicators(ctx, candleData, priceRange, padding, chartWidth, chartHeight);
        
        // Рисуем оси
        ctx.fillStyle = '#ffffff';
        ctx.font = '12px Arial';
        
        // Ось Y (цены)
        for (let i = 0; i <= 5; i++) {
            const price = priceRange.min + (i / 5) * (priceRange.max - priceRange.min);
            const y = padding + chartHeight - (i / 5) * chartHeight;
            ctx.textAlign = 'left';
            ctx.fillText(price.toFixed(2), canvas.width - padding + 5, y + 4);
        }
        
        // Ось X (время)
        for (let i = 0; i <= 5; i++) {
            const index = Math.floor((i / 5) * (candleData.length - 1));
            const candle = candleData[index];
            const x = padding + (i / 5) * chartWidth;
            ctx.textAlign = 'center';
            ctx.fillText(
                new Date(candle.timestamp).toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'}),
                x,
                canvas.height - 5
            );
        }
    }

    // Отрисовка индикаторов сделок
    drawTradeIndicators(ctx, candleData, priceRange, padding, chartWidth, chartHeight) {
        this.trades.forEach(trade => {
            if (trade.pair !== this.currentPair) return;
            
            // Находим ближайшую свечу по времени
            let closestIndex = 0;
            let minTimeDiff = Math.abs(candleData[0].timestamp - trade.timestamp);
            
            candleData.forEach((candle, index) => {
                const timeDiff = Math.abs(candle.timestamp - trade.timestamp);
                if (timeDiff < minTimeDiff) {
                    minTimeDiff = timeDiff;
                    closestIndex = index;
                }
            });
            
            const x = padding + closestIndex * (chartWidth / candleData.length) + (chartWidth / candleData.length) / 2;
            const y = padding + chartHeight - ((trade.price - priceRange.min) / (priceRange.max - priceRange.min)) * chartHeight;
            
            // Рисуем круг с пунктирной границей
            const isProfit = trade.profit > 0;
            ctx.strokeStyle = isProfit ? '#f6e05e' : '#90cdf4';
            ctx.fillStyle = isProfit ? 'rgba(255, 193, 7, 0.3)' : 'rgba(66, 153, 225, 0.3)';
            ctx.lineWidth = 2;
            ctx.setLineDash([4, 4]);
            
            ctx.beginPath();
            ctx.arc(x, y, 6, 0, 2 * Math.PI);
            ctx.fill();
            ctx.stroke();
            
            ctx.setLineDash([]);
        });
    }

    // Отрисовка графика объемов
    drawVolumeChart() {
        const canvas = this.elements.volumeChart;
        const ctx = canvas.getContext('2d');
        const candleData = this.candleData[this.currentPair] || [];
        
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        if (candleData.length === 0) return;
        
        const padding = 20;
        const chartWidth = canvas.width - padding * 2;
        const chartHeight = canvas.height - padding * 2;
        
        // Находим максимальный объем
        const maxVolume = Math.max(...candleData.map(candle => candle.volume));
        
        // Рисуем фон
        ctx.fillStyle = 'rgba(26, 32, 44, 0.8)';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        
        // Рисуем столбцы объемов
        const barWidth = chartWidth / candleData.length * 0.8;
        const barSpacing = chartWidth / candleData.length;
        
        candleData.forEach((candle, index) => {
            const x = padding + index * barSpacing + barSpacing / 2;
            const barHeight = (candle.volume / maxVolume) * chartHeight;
            const y = canvas.height - padding - barHeight;
            
            const isGreen = candle.close >= candle.open;
            ctx.fillStyle = isGreen ? 'rgba(56, 161, 105, 0.6)' : 'rgba(229, 62, 62, 0.6)';
            
            ctx.fillRect(x - barWidth / 2, y, barWidth, barHeight);
        });
    }

    // Открытие модального окна сделки
    openTradeModal(e) {
        const rect = this.elements.candlestickChart.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        
        // Вычисляем цену по позиции клика
        const candleData = this.candleData[this.currentPair] || [];
        if (candleData.length === 0) return;
        
        const priceRange = this.getPriceRange(candleData);
        const price = priceRange.max - ((y / this.elements.candlestickChart.height) * (priceRange.max - priceRange.min));
        
        // Заполняем форму
        this.elements.tradePair.innerHTML = '';
        Array.from(this.selectedPairs).forEach(pair => {
            const option = document.createElement('option');
            option.value = pair;
            option.textContent = this.formatPairName(pair);
            option.selected = pair === this.currentPair;
            this.elements.tradePair.appendChild(option);
        });
        
        this.elements.tradePrice.value = price.toFixed(2);
        this.elements.tradeAmount.value = '1.0';
        
        this.elements.tradeModal.style.display = 'block';
    }

    // Закрытие модального окна сделки
    closeTradeModal() {
        this.elements.tradeModal.style.display = 'none';
    }

    // Выполнение сделки
    executeTrade() {
        const type = this.elements.tradeType.value;
        const pair = this.elements.tradePair.value;
        const price = parseFloat(this.elements.tradePrice.value);
        const amount = parseFloat(this.elements.tradeAmount.value);
        
        if (!price || !amount) return;
        
        // Создаем новую сделку
        const trade = {
            id: Date.now(),
            timestamp: Date.now(),
            type: type,
            pair: pair,
            price: price,
            amount: amount,
            profit: 0 // Будет вычислено при закрытии сделки
        };
        
        // Для демонстрации сразу закрываем сделку с случайной прибылью
        const currentPrice = this.priceData[pair] || price;
        const priceChange = currentPrice - price;
        trade.profit = (type === 'buy' ? priceChange : -priceChange) * amount;
        
        this.trades.unshift(trade);
        
        this.updateTradesList();
        this.updateProfitCalculator();
        this.closeTradeModal();
        this.updateCharts();
    }

    // Обновление списка сделок
    updateTradesList() {
        this.elements.tradesList.innerHTML = '';
        
        this.trades.forEach(trade => {
            const tradeElement = document.createElement('div');
            tradeElement.className = `trade-item ${trade.profit >= 0 ? 'profit' : 'loss'}`;
            
            tradeElement.innerHTML = `
                <div class="trade-header">
                    <span>${this.formatPairName(trade.pair)}</span>
                    <span class="${trade.profit >= 0 ? 'positive' : 'negative'}">
                        ${trade.profit >= 0 ? '+' : ''}${trade.profit.toFixed(2)}
                    </span>
                </div>
                <div class="trade-details">
                    <span>Тип: ${trade.type === 'buy' ? 'Покупка' : 'Продажа'}</span>
                    <span>Цена: ${trade.price.toFixed(2)}</span>
                    <span>Объем: ${trade.amount.toFixed(4)}</span>
                    <span>Время: ${new Date(trade.timestamp).toLocaleTimeString()}</span>
                </div>
            `;
            
            this.elements.tradesList.appendChild(tradeElement);
        });
    }

    // Обновление калькулятора прибыли
    updateProfitCalculator() {
        const totalProfit = this.trades.reduce((sum, trade) => sum + trade.profit, 0);
        
        this.elements.profitAmount.textContent = `$${totalProfit.toFixed(2)}`;
        this.elements.profitAmount.className = `profit-amount ${totalProfit >= 0 ? 'positive' : 'negative'}`;
    }

    // Обновление интерфейса биржевого стакана
    updateOrderBookUI() {
        const book = this.orderBook[this.currentPair];
        if (!book) return;
        
        // Обновляем заголовок
        this.elements.orderBook.querySelector('h3').textContent = 
            `Биржевой стакан - ${this.formatPairName(this.currentPair)}`;
        
        // Очищаем содержимое
        this.elements.asks.innerHTML = '<div class="order-book-header"><span>Цена</span><span>Объем</span></div>';
        this.elements.bids.innerHTML = '';
        
        // Добавляем аски (в обратном порядке)
        for (let i = Math.min(10, book.asks.length) - 1; i >= 0; i--) {
            const ask = book.asks[i];
            const row = document.createElement('div');
            row.className = 'order-row ask-row';
            row.innerHTML = `<span>${ask.price.toFixed(2)}</span><span>${ask.volume.toFixed(4)}</span>`;
            this.elements.asks.appendChild(row);
        }
        
        // Обновляем спред
        if (book.asks.length > 0 && book.bids.length > 0) {
            const spread = book.asks[0].price - book.bids[0].price;
            this.elements.spread.innerHTML = `<span>Спред: $${spread.toFixed(2)}</span>`;
        }
        
        // Добавляем биды
        for (let i = 0; i < Math.min(10, book.bids.length); i++) {
            const bid = book.bids[i];
            const row = document.createElement('div');
            row.className = 'order-row bid-row';
            row.innerHTML = `<span>${bid.price.toFixed(2)}</span><span>${bid.volume.toFixed(4)}</span>`;
            this.elements.bids.appendChild(row);
        }
    }

    // Обновление всего интерфейса
    updateUI() {
        this.updateCharts();
        this.updateOrderBookUI();
    }

    // Генерация тестовых данных
    generateMockData() {
        // Генерируем исторические данные для каждой пары
        this.availablePairs.forEach(pair => {
            this.candleData[pair] = [];
            let basePrice;
            
            // Устанавливаем базовые цены
            switch (pair) {
                case 'BTCUSDT':
                    basePrice = 45000 + Math.random() * 20000;
                    break;
                case 'ETHUSDT':
                    basePrice = 3000 + Math.random() * 2000;
                    break;
                case 'LTCBTC':
                    basePrice = 0.001 + Math.random() * 0.0005;
                    break;
                default:
                    basePrice = 100 + Math.random() * 50;
            }
            
            const now = Date.now();
            const timeframeMs = this.getTimeframeMs(this.chartTimeframe);
            
            // Генерируем 100 свечей в прошлом
            for (let i = 100; i >= 0; i--) {
                const timestamp = now - i * timeframeMs;
                const open = basePrice;
                const change = (Math.random() - 0.5) * basePrice * 0.02;
                const close = Math.max(0, open + change);
                const high = Math.max(open, close) * (1 + Math.random() * 0.01);
                const low = Math.min(open, close) * (1 - Math.random() * 0.01);
                const volume = Math.random() * 100 + 10;
                
                this.candleData[pair].push({
                    timestamp,
                    open,
                    high,
                    low,
                    close,
                    volume
                });
                
                basePrice = close;
            }
            
            // Устанавливаем текущую цену
            this.priceData[pair] = basePrice;
            this.updateOrderBook(pair, basePrice);
        });
        
        // Генерируем несколько тестовых сделок
        for (let i = 0; i < 5; i++) {
            const pair = this.availablePairs[Math.floor(Math.random() * this.availablePairs.length)];
            const price = this.priceData[pair] * (0.95 + Math.random() * 0.1);
            const trade = {
                id: Date.now() + i,
                timestamp: Date.now() - Math.random() * 3600000,
                type: Math.random() > 0.5 ? 'buy' : 'sell',
                pair: pair,
                price: price,
                amount: Math.random() * 5 + 0.1,
                profit: (Math.random() - 0.5) * 1000
            };
            this.trades.push(trade);
        }
        
        this.updateTradesList();
        this.updateProfitCalculator();
    }

    // Перегенерация данных свечей при смене таймфрейма
    regenerateCandleData() {
        // Для демонстрации просто пересоздаем данные
        // В реальном приложении здесь была бы загрузка данных с сервера
        this.generateMockData();
    }

    // Основной цикл обновления
    startUpdateLoop() {
        setInterval(() => {
            if (this.isConnected) {
                // Обновляем цены для демонстрации
                this.availablePairs.forEach(pair => {
                    if (this.priceData[pair]) {
                        const change = (Math.random() - 0.5) * this.priceData[pair] * 0.001;
                        this.priceData[pair] = Math.max(0, this.priceData[pair] + change);
                        this.updateCandleData(pair, this.priceData[pair]);
                        this.updateOrderBook(pair, this.priceData[pair]);
                    }
                });
                this.updateUI();
            }
        }, 1000);
        
        // Обновление графиков
        const updateCharts = () => {
            this.updateCharts();
            requestAnimationFrame(updateCharts);
        };
        updateCharts();
    }
}

// Инициализация терминала при загрузке страницы
document.addEventListener('DOMContentLoaded', () => {
    window.terminal = new TradingTerminal();
});
