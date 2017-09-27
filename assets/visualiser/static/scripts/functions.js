function pathJoin(parts, sep){
	var separator = sep || '/';
	var replace   = new RegExp(separator+'{1,}', 'g');
	return parts.join(separator).replace(replace, separator);
}

function render_flame(renderer_path){
	$.ajax({
		url: location.origin + pathJoin(["/", renderer_path, "flame"]),
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
}
