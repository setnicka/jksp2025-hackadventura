const background = document.getElementById('background');
const imageSrc = 'rm_disturbed_nino.jpg'; // Replace with your image path
const imageSize = 100; // Same as img width in CSS
const screenHeight = window.innerHeight;
const rowCount = Math.ceil(screenHeight / imageSize);
const screenWidth = window.innerWidth;

for (let i = 0; i < rowCount; i++) {
  const row = document.createElement('div');
  row.className = 'row';
  row.style.top = `${i * imageSize}px`;
  row.style.height = `${imageSize}px`;
  row.style.animationName = i % 2 === 0 ? 'scroll-left' : 'scroll-right';

  const imageCount = Math.ceil(screenWidth / imageSize) * 2; // repeat enough for scrolling
  for (let j = 0; j < imageCount; j++) {
    const img = document.createElement('img');
    img.src = imageSrc;
    row.appendChild(img);
  }

  background.appendChild(row);
}
