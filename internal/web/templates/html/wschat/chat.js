 document.addEventListener('DOMContentLoaded', () => {
      const form = document.getElementById('chat-form');
      const nameInput = document.getElementById('name');
      const messageInput = document.getElementById('message');
      const list = document.getElementById('messages');

      function formatTime(date) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      }

      function appendMessage({ author, text, mine }) {
        const wrap = document.createElement('div');
        wrap.className = `msg ${mine ? 'me' : 'them'}`;

        const content = document.createElement('div');
        content.textContent = text;
        wrap.appendChild(content);

        const meta = document.createElement('div');
        meta.className = 'meta';
        const nameEl = document.createElement('span');
        nameEl.textContent = author || 'Гость';
        const timeEl = document.createElement('span');
        timeEl.textContent = formatTime(new Date());
        meta.appendChild(nameEl);
        meta.appendChild(document.createTextNode('·'));
        meta.appendChild(timeEl);
        wrap.appendChild(meta);

        list.appendChild(wrap);
        list.scrollTop = list.scrollHeight;
      }

      function sendMessage() {
        const author = nameInput.value.trim() || 'Я';
        const text = messageInput.value.trim();
        if (!text) return;
        appendMessage({ author, text, mine: true });
        messageInput.value = '';
        messageInput.focus();
      }

      form.addEventListener('submit', (e) => {
        e.preventDefault();
        sendMessage();
      });

      // Отправка по Ctrl/Cmd+Enter, перенос строки по Enter
      messageInput.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
          e.preventDefault();
          sendMessage();
        }
      });

      // Демосообщение
      appendMessage({ author: 'Система', text: 'Добро пожаловать! Напишите сообщение ниже.', mine: false });
    });