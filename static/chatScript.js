const messageForm = document.getElementById('message-form');
const messageInput = document.getElementById('message-input');
const chatWindow = document.getElementById('chat-window');

function createBubbleForMessageFromUser(message, user, bubbleOwnerClass, bubbleFromClass){
  const bubble = document.createElement('div');
  bubble.classList.add('chat-bubble', bubbleOwnerClass, bubbleFromClass);
  bubble.innerHTML = `<p><strong>${user}</strong></p><p>${message}</p>`;
  chatWindow.appendChild(bubble);
  // Scroll to the bottom of the chat window when a new message is added
  chatWindow.scrollTop = chatWindow.scrollHeight;
}

// Get the current port number
var port = window.location.port;
var wsUrl = "ws://" + window.location.hostname + ":" + port + "/ws";
var socket = new WebSocket(wsUrl);

// const socket = new WebSocket('ws://localhost:8080/ws');

socket.addEventListener('message', function(event) {
  const message = event.data;
  // Handle incoming message from server
  console.log('Received message:', message);
  const bubble = document.createElement('div');
  const userName = message.split("|")[1]
  const content = message.split("|")[2]
  bubble.classList.add('chat-bubble', 'them-bubble', 'incoming-bubble');
  bubble.innerHTML = `<p><strong>${userName}</strong></p><p>${content}</p>`;
  chatWindow.appendChild(bubble);
  // Scroll to the bottom of the chat window when a new message is added
  chatWindow.scrollTop = chatWindow.scrollHeight;
});


messageForm.addEventListener('submit', (e) => {
    e.preventDefault();
    const text = messageInput.value.trim();

    // Deal with message input by user
    if (text !== '') {
      const data = {
        input_message: text
      };
      console.log(data)
      // Handle with sended message from client
      fetch('/send-message', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
      })
        .then(response => {
          if (!response.ok) {
            throw new Error('Failed to send message');
          }
          return response.json();
        })
        .then(data => {
          console.log('Response:', data);
          // // Handle the response from the backend here if necessary
        })
        .catch(error => {
          console.error(error);
        });

          
      // createBubbleForMessageFromUser(text, "You", 'you-bubble', 'outgoing-bubble')
      const bubble = document.createElement('div');
      bubble.classList.add('chat-bubble', 'you-bubble', 'outgoing-bubble');
      bubble.innerHTML = `<p><strong>You</strong></p><p>${text}</p>`;
      chatWindow.appendChild(bubble);
      // Scroll to the bottom of the chat window when a new message is added
      chatWindow.scrollTop = chatWindow.scrollHeight;
      messageInput.value = '';
    }


  });
