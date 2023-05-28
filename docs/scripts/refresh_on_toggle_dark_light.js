var paletteSwitcher1 = document.getElementById("__palette_1");
var paletteSwitcher2 = document.getElementById("__palette_2");

function changeSVGColor(color) {
  var svg = document.getElementById("color-change-svg").contentDocument;
  var elements = svg.getElementsByClassName("primaryColor");
  for (var i = 0; i < elements.length; i++) elements[i].style.fill = color;
}

paletteSwitcher1.addEventListener("change", function () {
  changeSVGColor("#FFCD28");
  location.reload();
});

paletteSwitcher2.addEventListener("change", function () {
  location.reload();
});
