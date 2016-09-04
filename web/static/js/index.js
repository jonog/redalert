
var redalert = new Vue({
  el: '#wrapper',
  data: {
    checks: [],
    failCount: null,
    dataResolved: false
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
            vm.checks.forEach(processCheck);
            vm.failCount = vm.checks.filter(function(c) { return c.status === 'FAILING'}).length;
            vm.dataResolved = true;
          })
          .catch(function (error) {
            console.log(error);
          })
      }
  },
  filters: {
    lowercase: function (value) {
      if (!value) return ''
      return value.toLowerCase()
    }
  }
})

function processCheck(check) {
  const lastEvent = _.first(check.events);
  if (lastEvent) {
    const metricNames = _.keys(lastEvent.data.metrics);
    check.selectedMetric = metricNames[0];
    check.selectedMetricValue = round(lastEvent.data.metrics[check.selectedMetric], 2);
    check.lastEvent = lastEvent;
    if (check.status === "FAILING") {
      check.errors = lastEvent.messages.join(', ');
    }
  } else {
    check.selectedMetric = null;
    check.selectedMetricValue = null;
  }

  let totalChecks = check.stats.failure_total + check.stats.successful_total;
  check.successRate = totalChecks > 0 ? round(100 * check.stats.successful_total / totalChecks, 2) : null;
  check.totalChecks = totalChecks;
  check.stateTransitionedAt = _.isNull(check.stats.state_transitioned_at) ? '' : timeAgo(new Date(check.stats.state_transitioned_at));
}

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
          },
          axisY: {
            labelInterpolationFnc: function(value) {
              return round(value, 2);
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

function round(value, decimals) {
  return Number(Math.round(value+'e'+decimals)+'e-'+decimals);
}

const timeUnits = [
  { name: "s", limit: 60, inSeconds: 1 },
  { name: "m", limit: 3600, inSeconds: 60 },
  { name: "h", limit: 86400, inSeconds: 3600  },
  { name: "d", limit: 604800, inSeconds: 86400 },
  { name: "w", limit: 2629743, inSeconds: 604800  },
  { name: "m", limit: 31556926, inSeconds: 2629743 },
  { name: "y", limit: null, inSeconds: 31556926 }
]

function timeAgo(target){
  var diff = (new Date() - target) / 1000;
  var i = 0, unit;
  while (unit = timeUnits[i++]) {
    if (diff < unit.limit || !unit.limit){
      var diff =  Math.floor(diff / unit.inSeconds);
      return diff + unit.name;
    }
  };
}
