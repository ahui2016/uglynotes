const export_btn = $('#export');
const json_btn = $('#json');

export_btn.click(() => {
    ajaxGet('/api/backup/export', export_btn, () => {
        // onSuccess
        export_btn.hide();
        json_btn.show();
    })
});