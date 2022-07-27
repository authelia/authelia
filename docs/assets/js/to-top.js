var topbutton = document.getElementById('toTop');

if (topbutton !== null) {

  topbutton.style.display = 'none';
  window.onscroll = function() {
    scrollFunction()
  };

}

function scrollFunction() {

  if (document.body.scrollTop > 40 || document.documentElement.scrollTop > 40) {
    topbutton.style.display = 'block';
  } else {
    topbutton.style.display = 'none';
  }

}
