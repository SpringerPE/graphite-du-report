$(document).ready(function(){
	var someb64data = ''
	var svgDiv = $('#svgdiv')

	$.ajax({
	    url: location.origin + "/flame_image",
	    async: true,
	    dataType: "html",
      	    beforeSend: function() {
      		$('#loaderImg').show();    /*showing  a div with spinning image */
        	    },
	    success: function(data){
       		/*  simply hide the image */    
       		$('#loaderImg').hide();
       		/*  your code here   */
	        	$('#svgdiv').append(data);
	    },

	});
});
