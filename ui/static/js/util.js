// getTemplateRefs returns a key value pair list of all sub-elements that contain a ref attribute.
function getTemplateRefs(element) {
    let refs = {};
    element.querySelectorAll("[ref]").forEach(function (ref, index) {
        refs[ref.getAttribute("ref")] = ref;
    });
    return refs
}