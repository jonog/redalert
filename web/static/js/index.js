
var redalert = new Vue({
  el: '#checks',
  data: {
    checks: []
  },
  created: function() {
    var vm = this;
    setInterval(function(){ vm.getChecks(); }, 3000);
  },
  methods: {
    getChecks: function () {
        var vm = this;
        axios.get(window.baseURL + '/v1/stats')
          .then(function (response) {
            vm.checks = response.data;
            vm.checks.forEach(function(check) {
              const lastEvent = _.last(check.events);
              if (lastEvent) {
                const metricNames = _.keys(lastEvent.data.metrics);
                // TODO: make metric selectable
                check.selectedMetric = metricNames[0];
              } else {
                check.selectedMetric = null;
              }
            })
          })
          .catch(function (error) {
            console.log(error);
          })
      }
  }
})

Vue.component('chartist', {
  props: ['metric', 'data'],
  template: '<div class="check-chart"></div>',
  mounted: function() {
    this.draw();
  },
  methods: {
    draw: function() {
      var vm = this;
      if (!vm.metric) {
        return
      }
      if (!vm.data || vm.data.length === 0) {
        return
      }
      const chart = new Chartist.Line(vm.$el, { series: generateSeries(vm.data, vm.metric) },
        {
          series: generateSeriesOpts(vm.metric),
          axisX: {
              type: Chartist.FixedScaleAxis,
              divisor: 6,
              labelInterpolationFnc: function(value) {
                return moment(value).format('HH:mm');
              }
            }
        },
        [
          ['screen', {
            showPoint: false
          }]
        ]);
    }
  },
  watch: {
    'data': {
      handler: 'draw',
      deep: true
    }
  }
})

function generateSeries(events, metricName) {
  var series = [];
  series.push({
    name: metricName,
    data: events.map(function(e) {
      return {
        x: new Date(e.time),
        y: e.data.metrics[metricName]
      };
    })
  });
  return series;
}

function generateSeriesOpts(metricName) {
  var seriesOpts = {}
  seriesOpts[metricName] = {
    lineSmooth: Chartist.Interpolation.step({
      postpone: true,
      fillHoles: false
    })
  };
  return seriesOpts;
}
