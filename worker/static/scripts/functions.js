$(document).ready(function(){
	$.ajax({
	    url: location.origin + "/render/flame",
	    async: true,
	    dataType: "html",

		beforeSend: function(xhr) {
			/*showing  a div with spinning image */
      		$('#loaderImg').show();
			xhr.setRequestHeader('X-Cache-Enabled', 'true');
		},
	    success: function(data){
       		/*  simply hide the image */    
       		$('#loaderImg').hide();
			$('#svgdiv').append(data);
	    }
	});
});
