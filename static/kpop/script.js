const background = document.getElementById('background');
const imageSrc = background.title; // Replace with your image path
const rowHeight = 100; // Same as img width in CSS
const screenHeight = window.innerHeight;
const rowCount = Math.ceil(screenHeight / rowHeight);

for (let i = 0; i < rowCount; i++) {
  const row = document.createElement('div');
  row.className = 'row';
  row.style.top = `${i * rowHeight}px`;
  row.style.animationName = i % 2 === 0 ? 'scroll-left' : 'scroll-right';

  // Duplicate images to fill width twice
  const imagesPerRow = Math.ceil(window.innerWidth / rowHeight) * 2;
  for (let j = 0; j < imagesPerRow; j++) {
    const img = document.createElement('img');
    img.src = imageSrc;
    img.alt = '';
    row.appendChild(img);
  }

  background.appendChild(row);
}
